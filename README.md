Golang 单元测试实践
--------------------

每个严谨的项目都应该有单元测试及早发现程序中的问题，以及保障程序现在和未来的正确性。我们新加入一个项目的时候，常常在被要求熟悉代码的同时给现有代码加一些单元测试；或者自己的代码写到一定程度后，希望加一些单元测试看看有没有问题。这时往往会发现没法在不改动现有代码的情况下添加单元测试，这就引出一个很尴尬的问题~ 不是所有代码都是可测试的。

比如以下例子：

```
func AddPerson(name string) error {
    db, _ := sqlx.Open("mysql", "...dsn...")
    _, err := db.Exec("INSERT INTO person (name) VALUES (?)", name)
    return err
}
```

在函数中写死了mysql的连接方式，硬要写单元测试的话，会污染生产环境的数据库。

还有其它一些情况，比如一个函数从很多外部依赖获取数据并处理，函数的输入输出过于复杂。

一般来说，没法测试的代码都是不太好的代码，它们往往没有合理组织，不灵活甚至错误百出。直接说明怎样的代码可方便测试有点难，但我们可以通过看看各种情况下可以怎样合理单元测试，反推出应该怎样写出方便测试的代码。

本文主要说明 Golang 单元测试用到的工具以及一些方法，包括：

* 使用 Table Driven 的方式写测试代码
* 使用 testify/assert 简化条件判断
* 使用 testify/mock 隔离第三方依赖或者复杂调用
* mock http request
* stub redis
* stub MySQL

### 使用 Table Driven 的方式写测试代码

测试一个 routine 分几个步骤：准备数据，调用routine，判断返回。而且需要测试不同的情况。如果每种情况都手工写一次代码会很繁琐，可以使用 Table Driven 的方式让测试代码看起来简洁易懂。

比如有一个要测试一个取模运算 routine：

```
func Mod(a, b int) (r int, err error) {
    if b == 0 {
        return 0, fmt.Errorf("mod by zero")
    }
    return a%b, nil
}
```

可以这样测试：

```
func TestMod(t *testing.T) {
    tests := []struct {
        a int
        b int
        r int
        hasErr bool
    }{
        {a: 42, b: 9, r: 6, hasErr: false},
        {a: -1, b: 9, r: 8, hasErr: false},
        {a: -1, b: -9, r: -1, hasErr: false},
        {a: 42, b: 0, r: 0, hasErr: true},
    }

    for row, test := range tests {
        r, err := Mod(test.a, test.b)
        if test.hasError {
            if err == nil {
                t.Errorf("should have error, row: %d", row)
            }
            continue
        }
        if err != nil {
            t.Errorf("should not have error, row: %d", row)
        }
        if r != test.r {
            t.Errorf("r is expected to be %d but now %d, row: %d", test.r, r, row)
        }
    }
}
```

这样，以后如果想到新的边缘情况，就可以很方便地添加测试用例。

### 使用 testify/assert 简化条件判断

上面例子中很多 if xxx { t.Errorf(...) } 的代码，看上去复杂，语义不清晰。我们可以使用 github.com/stretchr/testify 的 assert 简化这些代码。上面的 for 循环可以简化成下面这样：

```
import github.com/stretchr/testify/assert"

for row, test := range tests {
    r, err := Mod(test.a, test.b)
    if test.hasError {
        assert.Error(t, err, "row %d", row)
        continue
    }
    assert.NoError(t, err, "row %d", row)
    assert.Equal(t, test.r, r, "row %d", row)
}
```

除了 Equal Error NoError，assert 还提供很多其它意义明确的判断方法比如：NotNil, NotEmpty, HTTPSucess 等。

### 使用 testify/mock 隔离第三方依赖或者复杂调用

很多时候测试环境很难具备 routine 执行的必要条件。比如要查询 consul 里的 KV，即使准备了测试consul，也要往里面先塞测试数据，十分麻烦。又比如查询 AWS S3 的文件列表，每个开发人员一个测试 bucket 太混乱，大家用同一个测试bucket 更混乱。必须找个方式伪造 consul client 和 AWS S3 client，通过伪造 client 查询 KV 或者 bucket 时不用真的去连 consul 和 AWS， 直接返回预设的结果。

首先要考虑怎样伪造一个 client，假设 client 被定义为 var client *SomeClient。当 SomeClient 的定义是 type SomeClient struct{...}，时我们永远没法在 test 环境修改 client 的行为。当 type SomeClient interface{...} 时，我们可以在测试代码中实现一个符合 SomeClient interface 的 struct，用这个 struct 的实例替换原来的 client。

假设一个IP限流程序从 consul 获取阈值并更新：

```
type SettingGetter interface {
    Get(key string) ([]byte, error)
}

type ConsulKV struct {
    kv *consul.KV
}

func (ck *ConsulKV) Get(key string) (value []byte, err error) {
    pair, _, err := ck.kv.Get(key, nil)
    if err != nil {
        return nil, err
    }
    return pair.Value, nil
}

type IPLimit struct {
    Threshold     int64
    SettingGetter SettingGetter
}

func (il *IPLimit) UpdateThreshold() error {
    value, err := il.SettingGetter.Get(KeyIPRateThreshold)
    if err != nil {
        return err
    }

    threshold, err := strconv.Atoi(string(value))
    if err != nil {
        return err
    }

    il.Threshold = int64(threshold)
    return nil
}
```

因为 consul.KV 是个 struct，没法方便替换，而我们只用到它的 Get 功能，所以可以简单定义一个 SettingGetter，并在 ConsulKV 实现这个接口，IPLimit 通过 SettingGetter 获得值，转换并更新。

在测试的时候，我们不能使用 ConsulKV，需要伪造一个 SettingGetter，像下面这样：

```
type MockSettingGetter struct {}

func (m *MockSettingGetter) Get(key string) ([]byte, error) {
    if key == "threshold" {
        return []byte("100"), nil
    }
    if key == "nothing" {
        return nil, fmt.Errorf("notfound")
    }
    ...
}

ipLimit := &IPLimit{SettingGetter: &MockSettingGetter{}}
// ... test with ipLimit
```

这样的确可以隔离 test 对 consul 的访问，但不方便 Table Driven。可以使用 testfiy/mock 改造一下，变成下面这样子：

```
import "github.com/stretchr/testify/mock"

type MockSettingGetter struct {
    mock.Mock
}

func (m *MockSettingGetter) Get(key string) (value []byte, err error) {
    args := m.Called(key)
    return args.Get(0).([]byte), args.Error(1)
}

func TestUpdateThreshold(t *testing.T) {
    tests := []struct {
        v      string
        err    error
        rs     int64
        hasErr bool
    }{
        {v: "1000", err: nil, rs: 1000, hasErr: false},
        {v: "a", err: nil, rs: 0, hasErr: true},
        {v: "", err: fmt.Errorf("consul is down"), rs: 0, hasErr: true},
    }

    for idx, test := range tests {
        mockSettingGetter := new(MockSettingGetter)
        mockSettingGetter.On("Get", mock.Anything).Return([]byte(test.v), test.err)

        limiter := &IPLimit{SettingGetter: mockSettingGetter}
        err := limiter.UpdateThreshold()
        if test.hasErr {
            assert.Error(t, err, "row %d", idx)
        } else {
            assert.NoError(t, err, "row %d", idx)
        }
        assert.Equal(t, test.rs, limiter.Threshold, "thredshold should equal, row %d", idx)
    }
}
```

testfiy/mock 使得伪造对象的输入输出值可以在运行时决定。更多技巧可看 testify/mock 的文档。

再说到上面提到的 AWS S3，AWS 的 Go SDK 已经给我们定义好 API 的 interface 了，每个服务下都有个 xxxiface 目录，比如 S3 的是 github.com/aws/aws-sdk-go/service/s3/s3iface，如果查看它的源码，会发现它的 API interface 有一大堆方法，将这几十个方法都伪造一次而只用其中一两个显得很蠢。一个方法是将 S3 的 API 像上面那样再封装一下，另一个方法可以像下面这样：

```
import (
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type MockS3API struct {
    s3iface.S3API
    mock.Mock
}

func (m *MockS3API) ListObjects(input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
    args := m.Called(input)
    return args.Get(0).(*s3.ListObjectsOutput), args.Error(1)
}
```

struct 里内嵌一个匿名 interface，免去定义无关方法的苦恼。

### mock http request

单元测试中还有个难题是怎样伪造 HTTP 请求的结果，如果像上面那样封装一下，很可能会漏掉一些极端情况的测试，比如没法连接网络，失败的状态码。Golang 有个 httptest 库，可以在 test 的时候创建一个 server，让 client 连上 server 测试。这样做会有点绕，事实上 Golang 的 http.Client 有个 Transport 成员，输入输出都通过它，可以通过篡改 Transport 返回我们需要的数据。

以一段获得本机公有IP的代码为例：

```
type IPApi struct {
    Client *http.Client
}

// MyIP return public ip address of current machine
func (ia *IPApi) MyIP() (ip string, err error) {
    resp, err := ia.Client.Get(MyIPUrl)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    if resp.StatusCode != 200 {
        return "", fmt.Errorf("status code: %d", resp.StatusCode)
    }

    infos := make(map[string]string)
    err = json.Unmarshal(body, &infos)
    if err != nil {
        return "", err
    }

    ip, ok := infos["ip"]
    if !ok {
        return "", ErrInvalidRespResult
    }
    return ip, nil
}
```

可以这样写单元测试：

```
// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
    return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
    return &http.Client{
        Transport: RoundTripFunc(fn),
    }
}

func TestMyIP(t *testing.T) {
    tests := []struct {
        code     int
        text     string
        ip       string
        hasError bool
    }{
        {code: 200, text: "{\"ip\":\"1.2.3.4\"}", ip: "1.2.3.4", hasError: false},
        {code: 403, text: "", ip: "", hasError: true},
        {code: 200, text: "abcd", ip: "", hasError: true},
    }

    for row, test := range tests {
        client := NewTestClient(func(req *http.Request) *http.Response {
            assert.Equal(t, req.URL.String(), MyIPUrl, "ip url should match, row %d", row)
            return &http.Response{
                StatusCode: test.code,
                Body:       ioutil.NopCloser(bytes.NewBufferString(test.text)),
                Header:     make(http.Header),
            }
        })
        api := &IPApi{Client: client}

        ip, err := api.MyIP()
        if test.hasError {
            assert.Error(t, err, "row %d", row)
        } else {
            assert.NoError(t, err, "row %d", row)
        }
        assert.Equal(t, test.ip, ip, "ip should equal, row %d", row)
    }
}
```

### stub redis

假如程序里用到 Redis，要伪造一个 Redis Client 用之前的办法也是可以的。但因为有 miniredis 的存在，我们有更好的办法。miniredis 是在 Golang 程序中运行的 Redis Server，它实现了大部分原装 Redis 的功能，测试的时候 miniredis.Run() 然后将 Redis Client 连向 miniredis 就可以。

这种方式成为 stub，和 mock 有一些微妙的差别，可参考 [stackoverflow](https://stackoverflow.com/questions/3459287/whats-the-difference-between-a-mock-stub) 的讨论。

miniredis 具体使用方式如下：

```
var testRdsSrv *miniredis.Miniredis

func TestMain(m *testing.M) {
    s, err := miniredis.Run()
    if err != nil {
        panic(err)
    }
    defer s.Close()
    os.Exit(m.Run()
}

func TestSomeRedis(t *testing.T) {
    tests := []struct {...}{...}
    for row, test := range tests {
        testRdsSrv.FlushAll()
        rClient := redis.NewClient(&redis.Options{
            Addr: testRdsSrv.Addr(),
        })
        // do something with rClient
    }
    testRdsSrv.FlushAll()
}
```

主要需要考虑的是保障每个测试都有个干净的 redis 数据库。

### stub MySQL

要测试用到关系数据库的代码更加麻烦，因为很多时候看程序正确与否就看它写入到数据库里的数据对不对，关系数据库的操作不能简单 mock 一下，测试的时候需要一个真的数据库来查看程序对不对。

MySQL 或者其它关系数据库没有类似 miniredis 的解决方案，需要我们在测试之前搭好一个干净的测试 MySQL Server，并且里面的表也要建好。这些条件没法单靠写 Go 代码实现，需要使用一些工具以及在代码工程里做一点约定。

我想到的一个方案是，工程里有个 sql 文件，里面有建库建表语句，编写一个 docker-compose 配置用于创建 MySQL Server，执行建库建表语句，编写 Makefile 将启动 MySQL，建表，go test，关闭 MySQL 组织起来。

我试了一下，实现了整个流程后测试挺顺畅的，相关配置代码太多就不在这里贴了，有兴趣可看 [Github testingo](https://github.com/euclidr/testingo)

实现过程中主要遇到两个问题，一个是需要确认 MySQL 的 docker 真正正常运行后才能建库建表，一个是默认 storage-engine 是 Innodb 可以考虑修改为 Memory 加快测试的速度。

## 参考资料

1. [以上所有测试的详细例子](https://github.com/euclidr/testingo)
2. [testing](https://golang.org/pkg/testing/)
3. [testify](https://github.com/stretchr/testify)
4. [Unit Testing http client in Go](http://hassansin.github.io/Unit-Testing-http-client-in-Go)
5. [Integration Test With Database in Golang]https://hackernoon.com/integration-test-with-database-in-golang-355dc123fdc9
6. [miniredis](https://github.com/alicebob/miniredis)