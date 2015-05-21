/*
 Source Server         : Local Sever
 Source Server Type    : MariaDB
 Source Server Version : 100017
 Source Host           : localhost
 Source Database       : yayoi

 Target Server Type    : MariaDB
 Target Server Version : 100017
 File Encoding         : utf-8

 Date: 05/20/2015 19:19:08 PM
*/

SET NAMES utf8;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
--  Table structure for `authentications`
-- ----------------------------
DROP TABLE IF EXISTS `authentications`;
CREATE TABLE `authentications` (
  `Token` varchar(30) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `UserID` bigint(20) unsigned DEFAULT NULL,
  `Ip` varchar(41) DEFAULT NULL,
  `Time` bigint(20) DEFAULT NULL,
  `Expires` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`Token`)
) ENGINE=Aria DEFAULT CHARSET=utf8 PAGE_CHECKSUM=1;

-- ----------------------------
--  Table structure for `login`
-- ----------------------------
DROP TABLE IF EXISTS `login`;
CREATE TABLE `login` (
  `Ip` varchar(41) NOT NULL,
  `LoginNonce` binary(32) DEFAULT NULL,
  `LoginAttempts` int(11) DEFAULT NULL,
  `LastAttempt` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`Ip`)
) ENGINE=Aria DEFAULT CHARSET=utf8 PAGE_CHECKSUM=1;

-- ----------------------------
--  Table structure for `settings`
-- ----------------------------
DROP TABLE IF EXISTS `settings`;
CREATE TABLE `settings` (
  `Name` varchar(255) NOT NULL,
  `Value` text,
  PRIMARY KEY (`Name`),
  UNIQUE KEY `setting_name` (`Name`)
) ENGINE=Aria DEFAULT CHARSET=utf8 PAGE_CHECKSUM=1;

-- ----------------------------
--  Table structure for `tags`
-- ----------------------------
DROP TABLE IF EXISTS `tags`;
CREATE TABLE `tags` (
  `Id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `Value` text NOT NULL,
  `Alias` text NOT NULL,
  `UseCount` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`Id`)
) ENGINE=Aria AUTO_INCREMENT=4 DEFAULT CHARSET=utf8 PAGE_CHECKSUM=1;

-- ----------------------------
--  Table structure for `uploads`
-- ----------------------------
DROP TABLE IF EXISTS `uploads`;
CREATE TABLE `uploads` (
  `UserID` bigint(20) unsigned DEFAULT NULL,
  `MD5` varchar(32) DEFAULT NULL,
  `SHA1` varchar(40) DEFAULT NULL,
  `SHA256` varchar(64) NOT NULL,
  `SHA512` varchar(128) DEFAULT NULL,
  `Extension` varchar(5) DEFAULT NULL,
  `FileSize` bigint(20) DEFAULT NULL,
  `Width` int(11) DEFAULT NULL,
  `Height` int(11) DEFAULT NULL,
  `ThumbnailExtension` varchar(5) DEFAULT NULL,
  `ThumbnailFileSize` bigint(20) DEFAULT NULL,
  `ThumnailWidth` int(11) DEFAULT NULL,
  `ThumnailHeight` int(11) DEFAULT NULL,
  `Rating` varchar(1) DEFAULT '',
  `SourceURL` text,
  `Author` varchar(50) DEFAULT '',
  `AuthorURL` text,
  `Tags` text,
  `Time` bigint(20) DEFAULT NULL,
  `Submitted` int(11) DEFAULT '0',
  PRIMARY KEY (`SHA256`)
) ENGINE=Aria DEFAULT CHARSET=utf8 PAGE_CHECKSUM=1;

-- ----------------------------
--  Table structure for `users`
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `Id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `Name` varchar(50) DEFAULT NULL,
  `Email` varchar(100) DEFAULT NULL,
  `Password` binary(64) DEFAULT NULL,
  `PasswordSalt` binary(32) DEFAULT NULL,
  `ResetKey` varchar(30) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,
  `SignupKey` varchar(30) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,
  `ApiKey` varchar(30) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,
  `Level` int(11) DEFAULT NULL,
  `JoinTime` bigint(20) DEFAULT NULL,
  `LastLoginTime` bigint(20) DEFAULT NULL,
  `ResetRequestTime` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`Id`)
) ENGINE=Aria AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 PAGE_CHECKSUM=1;

SET FOREIGN_KEY_CHECKS = 1;
