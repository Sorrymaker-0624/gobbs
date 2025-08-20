# 数据库表结构设计

本文档记录了 GoBBS 项目的数据库表结构信息。

## 1. 用户表 (`user`)

用于存储所有用户的基础信息。

| 字段名        | 数据类型            |    约束/备注        |
|:-------------|:------------------|:----------------|
| `id`         | `BIGINT UNSIGNED` | 主键, 自增          |
| `user_id`    | `BIGINT UNSIGNED` | 业务主键, 唯一, 非空    |
| `username`   | `VARCHAR(64)`     | 用户名, 唯一, 非空     |
| `password`   | `VARCHAR(255)`    | 密码 (存储哈希值), 非空  |
| `email`      | `VARCHAR(64)`     | 邮箱, 唯一          |
| `created_at` | `TIMESTAMP`       | 创建时间 (GORM自动管理) |
| `updated_at` | `TIMESTAMP`       | 更新时间 (GORM自动管理) |

## 2. 帖子表 (`post`)

用于存储用户发布的帖子。

| 字段名         | 数据类型          | 约束/备注                             |
| :------------- | :---------------- | :------------------------------------ |
| `id`           | `BIGINT UNSIGNED`   | 主键, 自增                          |
| `post_id`      | `BIGINT UNSIGNED`   | 业务主键, 唯一, 非空                  |
| `author_id`    | `BIGINT UNSIGNED`   | 作者ID (外键关联 user.user_id), 非空 |
| `community_id` | `BIGINT UNSIGNED`   | 社区/板块ID, 非空                     |
| `status`       | `TINYINT UNSIGNED`  | 帖子状态 (1:正常, 2:待审核), 默认1  |
| `title`        | `VARCHAR(255)`    | 帖子标题, 非空                        |
| `content`      | `LONGTEXT`        | 帖子正文, 非空                        |
| `created_at`   | `TIMESTAMP`       | 创建时间 (GORM自动管理)               |
| `updated_at`   | `TIMESTAMP`       | 更新时间 (GORM自动管理)               |