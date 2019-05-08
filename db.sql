CREATE TABLE `heartbeat` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user` varchar(255) NOT NULL,
  `time` datetime NOT NULL,
  `entity` varchar(1024) DEFAULT NULL,
  `type` varchar(255) NOT NULL,
  `category` varchar(255) DEFAULT NULL,
  `is_write` tinyint(4) NOT NULL,
  `branch` varchar(255) DEFAULT NULL,
  `language` varchar(255) DEFAULT NULL,
  `project` varchar(255) DEFAULT NULL,
  `operating_system` varchar(45) DEFAULT NULL,
  `editor` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE `user` (
  `user_id` varchar(255) NOT NULL,
  `api_key` varchar(255) NOT NULL,
  PRIMARY KEY (`user_id`),
  KEY `IDX_API_KEY` (`api_key`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;