/*
 Source Server         : Local Server
 Source Server Type    : MySQL
 Source Server Version : 50505
 Source Host           : 127.0.0.1
 Source Database       : yayoi

 Target Server Type    : MySQL
 Target Server Version : 50505
 File Encoding         : utf-8

 Date: 05/16/2015 19:12:29 PM
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
  `Ip` varchar(39) DEFAULT NULL,
  `Time` bigint(20) DEFAULT NULL,
  `Expires` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`Token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
--  Table structure for `login`
-- ----------------------------
DROP TABLE IF EXISTS `login`;
CREATE TABLE `login` (
  `Ip` varchar(39) NOT NULL,
  `LoginNonce` binary(32) DEFAULT NULL,
  `LoginAttempts` int(11) DEFAULT NULL,
  `LastAttempt` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`Ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

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
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;
