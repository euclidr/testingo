CREATE DATABASE `testingo`;

USE `testingo`;

CREATE TABLE `animal` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) CHARACTER SET utf8 NOT NULL,
  `place` varchar(100) CHARACTER SET utf8 NOT NULL,
  PRIMARY KEY (`id`)
);
