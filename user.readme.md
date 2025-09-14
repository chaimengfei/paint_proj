# 用户系统设计文档

## 概述

本系统采用统一的用户表设计，支持微信小程序和后台管理系统两套系统共用一条用户记录。通过手机号作为唯一标识，实现用户在不同平台间的数据关联。

## 设计理念

### 核心思想
- **统一用户表**：微信小程序用户和后台管理系统用户使用同一张 `user` 表
- **手机号唯一标识**：以手机号作为用户的唯一标识，确保数据一致性
- **灵活显示名称**：支持不同场景下的用户名称显示
- **来源追踪**：记录用户来源，便于数据分析和业务管理

### 业务场景
1. **线下客户**：管理员在后台添加客户信息，客户后续可能使用小程序
2. **线上客户**：用户直接通过小程序注册，后续可能被管理员管理
3. **混合客户**：先后台添加，后小程序绑定，或先小程序注册，后后台管理

## 数据库设计

### 用户表结构

```sql
CREATE TABLE user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    openid VARCHAR(100) COMMENT '微信OpenID',
    nickname VARCHAR(100) COMMENT '微信昵称（原始）',
    avatar VARCHAR(500) COMMENT '头像',
    mobile_phone VARCHAR(20) UNIQUE COMMENT '手机号（唯一标识）',
    source TINYINT NOT NULL DEFAULT 1 COMMENT '用户来源(1:小程序,2:后台添加,3:混合)',
    is_enable TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用(1:启用,0:禁用)',
    admin_display_name VARCHAR(100) COMMENT '后台管理系统显示的客户名称',
    wechat_display_name VARCHAR(100) COMMENT '微信小程序显示的客户名称',
    has_wechat_bind TINYINT NOT NULL DEFAULT 0 COMMENT '是否已绑定微信(1:是,0:否)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
);
```

### 字段说明

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `id` | BIGINT | 用户ID（主键） | 1, 2, 3... |
| `openid` | VARCHAR(100) | 微信OpenID | "oGZUI0..." |
| `nickname` | VARCHAR(100) | 微信昵称（原始） | "阳光男孩" |
| `avatar` | VARCHAR(500) | 头像URL | "https://..." |
| `mobile_phone` | VARCHAR(20) | 手机号（唯一标识） | "13800138000" |
| `source` | TINYINT | 用户来源 | 1:小程序, 2:后台添加, 3:混合 |
| `is_enable` | TINYINT | 是否启用 | 1:启用, 0:禁用 |
| `admin_display_name` | VARCHAR(100) | 后台显示名称 | "孙阳", "张总", "xx雕塑王平" |
| `wechat_display_name` | VARCHAR(100) | 微信显示名称 | "阳光男孩", "小张" |
| `has_wechat_bind` | TINYINT | 是否已绑定微信 | 1:是, 0:否 |

### 常量定义

```go
// 用户来源类型
const (
    UserSourceWechat = 1 // 小程序注册
    UserSourceAdmin  = 2 // 后台添加
    UserSourceMixed  = 3 // 混合（先后台添加，后小程序绑定）
)

// 用户状态
const (
    UserStatusDisabled = 0 // 禁用
    UserStatusEnabled  = 1 // 启用
)

// 微信绑定状态
const (
    WechatBindNo  = 0 // 未绑定微信
    WechatBindYes = 1 // 已绑定微信
)
```

## 业务流程

### 1. 后台添加用户流程

```
管理员操作：
1. 输入用户信息（admin_display_name, mobile_phone）
2. 系统检查手机号是否已存在
3. 如果不存在：创建用户记录
   - source = 2 (后台添加)
   - is_enable = 1 (启用)
   - has_wechat_bind = 0 (未绑定微信)
   - openid = NULL
4. 如果存在：提示用户已存在
```

**API接口：**
```
POST /admin/user/add
{
    "admin_display_name": "孙阳",
    "mobile_phone": "13800138000",
    "remark": "备注信息"
}
```

### 2. 小程序登录流程

```
用户操作：
1. 通过 code 换取 openid
2. 检查 openid 是否已存在
3. 如果存在：直接登录
4. 如果不存在：
   a. 创建新用户记录
   b. 设置 source = 1 (小程序)
   c. 设置 has_wechat_bind = 1
   d. 设置 wechat_display_name = nickname
```

**API接口：**
```
POST /api/user/login
{
    "code": "微信授权码",
    "nickname": "用户昵称",
    "avatar": "头像URL"
}
```

### 3. 小程序绑定手机号流程

```
用户操作：
1. 用户输入手机号
2. 系统查找该手机号对应的用户记录
3. 如果找到：
   a. 将当前微信用户绑定到现有用户
   b. 更新 openid 字段
   c. 设置 has_wechat_bind = 1
   d. 设置 source = 3 (混合)
4. 如果未找到：更新当前用户的手机号
```

**API接口：**
```
POST /api/user/bind-mobile
{
    "mobile_phone": "13800138000"
}
```

## 用户信息显示逻辑

### 后台管理系统显示

```go
func GetDisplayName(user *User) string {
    if user.AdminDisplayName != "" {
        return user.AdminDisplayName
    }
    if user.WechatDisplayName != "" {
        return user.WechatDisplayName
    }
    if user.Nickname != "" {
        return user.Nickname
    }
    return user.MobilePhone
}
```

**显示优先级：**
1. `admin_display_name`（如果存在）
2. `wechat_display_name`（如果存在且admin_display_name为空）
3. `nickname`（如果存在且上述都为空）
4. `mobile_phone`（如果上述都为空）

### 微信小程序显示

```go
func GetWechatDisplayName(user *User) string {
    if user.WechatDisplayName != "" {
        return user.WechatDisplayName
    }
    if user.Nickname != "" {
        return user.Nickname
    }
    if user.AdminDisplayName != "" {
        return user.AdminDisplayName
    }
    return user.MobilePhone
}
```

**显示优先级：**
1. `wechat_display_name`（如果存在）
2. `nickname`（如果存在且wechat_display_name为空）
3. `admin_display_name`（如果存在且上述都为空）
4. `mobile_phone`（如果上述都为空）

## API接口文档

### 后台用户管理接口

#### 1. 添加用户
```
POST /admin/user/add
Content-Type: application/json

{
    "admin_display_name": "孙阳",
    "mobile_phone": "13800138000",
    "remark": "备注信息"
}
```

**响应示例：**
```json
{
    "code": 0,
    "message": "添加用户成功",
    "data": {
        "id": 1,
        "admin_display_name": "孙阳",
        "mobile_phone": "13800138000",
        "source": 2,
        "is_enable": 1,
        "has_wechat_bind": 0,
        "created_at": "2024-01-15T10:30:00Z"
    }
}
```

#### 2. 获取用户列表
```
GET /admin/user/list?page=1&page_size=10&keyword=孙阳
```

**响应示例：**
```json
{
    "code": 0,
    "data": {
        "list": [
            {
                "id": 1,
                "admin_display_name": "孙阳",
                "mobile_phone": "13800138000",
                "source": 2,
                "is_enable": 1,
                "has_wechat_bind": 0,
                "created_at": "2024-01-15T10:30:00Z"
            }
        ],
        "total": 1,
        "page": 1,
        "page_size": 10
    }
}
```

#### 3. 根据ID获取用户
```
GET /admin/user/1
```

#### 4. 编辑用户
```
PUT /admin/user/edit
Content-Type: application/json

{
    "id": 1,
    "admin_display_name": "孙阳（更新）",
    "mobile_phone": "13800138000",
    "is_enable": 1
}
```

#### 5. 删除用户
```
DELETE /admin/user/del/1
```

### 小程序用户接口

#### 1. 用户登录
```
POST /api/user/login
Content-Type: application/json

{
    "code": "微信授权码",
    "nickname": "用户昵称",
    "avatar": "头像URL"
}
```

#### 2. 绑定手机号
```
POST /api/user/bind-mobile
Content-Type: application/json
Authorization: Bearer <token>

{
    "mobile_phone": "13800138000"
}
```

## 业务场景验证

### 场景1：后台添加用户，用户后续使用小程序

**步骤：**
1. 管理员在后台添加用户：
   ```json
   {
       "admin_display_name": "孙阳",
       "mobile_phone": "13800138000"
   }
   ```
2. 用户记录创建：
   ```json
   {
       "id": 1,
       "admin_display_name": "孙阳",
       "mobile_phone": "13800138000",
       "source": 2,
       "is_enable": 1,
       "has_wechat_bind": 0,
       "openid": null
   }
   ```
3. 用户在小程序登录并绑定手机号：
   ```json
   {
       "mobile_phone": "13800138000"
   }
   ```
4. 用户记录更新：
   ```json
   {
       "id": 1,
       "admin_display_name": "孙阳",
       "mobile_phone": "13800138000",
       "source": 3,
       "is_enable": 1,
       "has_wechat_bind": 1,
       "openid": "oGZUI0...",
       "wechat_display_name": "阳光男孩"
   }
   ```

**验证结果：**
- ✅ 后台显示：`admin_display_name` = "孙阳"
- ✅ 小程序显示：`wechat_display_name` = "阳光男孩"
- ✅ 数据关联：同一条用户记录
- ✅ 来源标识：`source` = 3 (混合)

### 场景2：用户直接小程序注册，后续后台管理

**步骤：**
1. 用户在小程序登录：
   ```json
   {
       "code": "微信授权码",
       "nickname": "阳光男孩"
   }
   ```
2. 用户记录创建：
   ```json
   {
       "id": 2,
       "openid": "oGZUI0...",
       "nickname": "阳光男孩",
       "source": 1,
       "is_enable": 1,
       "has_wechat_bind": 1,
       "wechat_display_name": "阳光男孩"
   }
   ```
3. 管理员在后台查看用户列表，可以看到该用户
4. 管理员可以编辑用户信息，添加 `admin_display_name`

**验证结果：**
- ✅ 小程序显示：`wechat_display_name` = "阳光男孩"
- ✅ 后台显示：`admin_display_name` = "阳光男孩"（管理员可编辑）
- ✅ 数据关联：同一条用户记录
- ✅ 来源标识：`source` = 1 (小程序)

### 场景3：出库选择客户

**步骤：**
1. 管理员在出库页面选择客户
2. 系统显示用户列表，按显示优先级排序
3. 管理员选择"孙阳"进行出库操作

**验证结果：**
- ✅ 显示名称：优先显示 `admin_display_name`
- ✅ 数据一致性：出库记录关联正确的用户ID
- ✅ 业务连续性：支持线上线下客户统一管理

## 数据迁移

### 现有用户数据处理

```sql
-- 为现有用户数据设置默认值
UPDATE user SET 
    source = 1,
    has_wechat_bind = CASE WHEN openid IS NOT NULL AND openid != '' THEN 1 ELSE 0 END,
    wechat_display_name = CASE WHEN nickname IS NOT NULL AND nickname != '' THEN nickname ELSE '' END
WHERE id > 0;
```

### 迁移脚本

```sql
-- 为用户表添加新字段
ALTER TABLE user 
    ADD COLUMN mobile_phone VARCHAR(20) UNIQUE COMMENT '手机号（唯一标识）' AFTER avatar,
    ADD COLUMN source TINYINT NOT NULL DEFAULT 1 COMMENT '用户来源(1:小程序,2:后台添加,3:混合)' AFTER mobile_phone,
    ADD COLUMN is_enable TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用(1:启用,0:禁用)' AFTER source,
    ADD COLUMN admin_display_name VARCHAR(100) COMMENT '后台管理系统显示的客户名称' AFTER is_enable,
    ADD COLUMN wechat_display_name VARCHAR(100) COMMENT '微信小程序显示的客户名称' AFTER admin_display_name,
    ADD COLUMN has_wechat_bind TINYINT NOT NULL DEFAULT 0 COMMENT '是否已绑定微信(1:是,0:否)' AFTER wechat_display_name;
```

## 注意事项

### 1. 数据一致性
- 手机号作为唯一标识，确保不重复
- 用户来源字段准确记录，便于数据分析
- 显示名称字段灵活使用，适应不同场景

### 2. 业务逻辑
- 后台添加用户时检查手机号唯一性
- 小程序绑定手机号时处理用户关联
- 用户信息更新时保持数据完整性

### 3. 扩展性
- 支持未来添加更多用户属性
- 支持更多用户来源类型
- 支持更复杂的用户关联逻辑

### 4. 安全性
- 手机号验证和格式检查
- 用户权限控制
- 数据访问日志记录

## 总结

本用户系统设计实现了微信小程序和后台管理系统的用户数据统一管理，通过手机号作为唯一标识，支持多种用户来源和绑定方式，提供了灵活的用户信息显示逻辑，满足了线上线下业务场景的需求。系统具有良好的扩展性和维护性，为后续业务发展提供了坚实的基础。
