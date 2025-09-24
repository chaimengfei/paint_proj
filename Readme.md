# 油漆销售系统

## 功能特性

### 店铺管理

系统支持多店铺管理，每个用户和商品都关联到特定店铺：

#### 店铺信息
- **燕郊店** (ID: 1): 河北省廊坊市三河市燕郊镇
- **涞水店** (ID: 2): 河北省保定市涞水县

#### 店铺分配逻辑
1. **小程序注册时**：
   - 如果用户提供了位置信息，系统会根据经纬度计算最近的店铺
   - 如果距离涞水店更近且距离小于50公里，分配涞水店
   - 否则默认分配燕郊店
   - 如果没有提供位置信息，默认分配燕郊店

2. **后台添加用户时**：
   - 管理员可以手动指定店铺ID
   - 如果不指定，默认分配燕郊店

#### 数据隔离
- 每个用户只能看到自己店铺的商品
- 购物车和订单都按店铺隔离
- 库存操作按店铺分别管理
- 入库操作需要管理员手动选择店铺

#### 店铺接口
系统提供店铺信息查询接口，支持小程序和后台管理系统：

**获取店铺列表**
```bash
# 小程序接口
curl -X GET "http://localhost:8009/api/shop/list"

# 后台管理接口
curl -X GET "http://localhost:8009/admin/shop/list"
```

**响应示例**
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "燕郊店",
      "code": "YJ001",
      "address": "河北省廊坊市三河市燕郊镇",
      "phone": "13161621688",
      "manager_name": "张三",
      "is_active": 1
    },
    {
      "id": 2,
      "name": "涞水店", 
      "code": "LS001",
      "address": "河北省保定市涞水县",
      "phone": "12345678910",
      "manager_name": "李四",
      "is_active": 1
    }
  ],
  "message": "获取店铺列表成功"
}
```

**注意事项**
- 店铺接口无需token验证，可直接访问
- 只返回启用状态（is_active=1）的店铺
- 支持小程序和后台管理系统同时使用

### 后台管理系统认证

系统支持后台管理员登录认证，使用 JWT Token 方案：

#### 管理员账号
- **root** (密码: admin123): 超级管理员，可操作所有店铺数据
- **lizengchun** (密码: lzc123): 燕郊店管理员，只能操作燕郊店数据
- **zhangweiyang** (密码: zwy123): 涞水店管理员，只能操作涞水店数据

#### 登录接口
```bash
curl -X POST "http://localhost:8009/admin/operator/login" \
  -H "Content-Type: application/json" \
  -d '{
    "operator_name": "lizengchun",
    "password": "your_password"
  }'
```

**响应示例**

**普通管理员登录响应**：
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "operator": {
      "id": 2,
      "operator_name": "lizengchun",
      "shop_id": 1,
      "real_name": "李增春",
      "phone": "131-0000-0000",
      "is_active": 1
    },
    "shop_info": {
      "id": 1,
      "name": "燕郊店",
      "code": "YJ001",
      "address": "河北省廊坊市三河市燕郊镇",
      "phone": "13161621688",
      "manager_name": "张三",
      "is_active": 1
    },
    "expires_in": 7200
  }
}
```

**超级管理员(root)登录响应**：
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "operator": {
      "id": 1,
      "operator_name": "root",
      "shop_id": 1,
      "real_name": "超级管理员",
      "phone": "400-000-0000",
      "is_active": 1
    },
    "shop_list": [
      {
        "id": 1,
        "name": "燕郊店",
        "code": "YJ001",
        "address": "河北省廊坊市三河市燕郊镇",
        "phone": "13161621688",
        "manager_name": "张三",
        "is_active": 1
      },
      {
        "id": 2,
        "name": "涞水店",
        "code": "LS001",
        "address": "河北省保定市涞水县",
        "phone": "0312-7654321",
        "manager_name": "张伟阳",
        "is_active": 1
      }
    ],
    "expires_in": 7200
  }
}
```

#### 权限控制
- **超级管理员 (root)**: 可以操作所有店铺的数据
- **普通管理员**: 只能操作自己所属店铺的数据
- **Token 有效期**: 2小时
- **自动店铺关联**: 所有操作自动关联到管理员所属店铺

#### 使用示例
```bash
# 添加商品（自动关联到管理员所属店铺）
curl -X POST "http://localhost:8009/admin/product/add" \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "新商品",
    "price": 100,
    "description": "商品描述"
  }'

# 批量入库（自动关联到管理员所属店铺）
curl -X POST "http://localhost:8009/admin/stock/batch/inbound" \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "product_id": 1,
        "quantity": 100
      }
    ]
  }'
```

#### 数据库初始化
首次使用前，需要在数据库中插入管理员数据：
```sql
-- 插入示例管理员数据（密码已加密）
-- 密码说明: root=admin123, lizengchun=lzc123, zhangweiyang=zwy123
INSERT INTO operator (name, password, shop_id, real_name, phone) VALUES
('root', '$2a$10$4YpHd00gQ7NuVkxHofK9Vupfm4rC/mwE0yfDtkoa0B/63Ec7uyTDG', 1, '超级管理员', '400-000-0000'),
('lizengchun', '$2a$10$rIzWQMbXpsFgQSSotodPDuVNKaphBIsYoxZrAb5orzrASOzH20MXW', 1, '李增春', '131-0000-0000'),
('zhangweiyang', '$2a$10$HLfwdIvwGjodaDkjQnrQVuhBnQsRytKtrvolXB861whv2n96.Lzge', 2, '张伟阳', '132-0000-0000');
```

**注意**: 管理员账号数量有限，如需新增管理员，请直接在数据库中插入记录。系统不提供通过接口创建、修改、删除管理员的功能。

### 关于登陆

1. 是否每次都要调用 GetOpenIDByCode?
**不是每次都要调用。标准流程如下：**


| 阶段             | 操作                                       | 调用 GetOpenIDByCode？ | 后端是否需要用它？ |
| ---------------- | ------------------------------------------ | ---------------------- | ------------------ |
| 第一次登录小程序 | wx.login() + 后端换取 openid + 绑定 userID | ✅ 是                   | ✅ 是               |
| 后续所有请求     | 带上 token 或 userID                       | ❌ 否                   | ❌ 不需要调用       |
| 发起支付         | 用存下来的 openid 发起支付下单             | ❌ 否                   | ✅ 用 openid        |

2. 微信小程序中，`openid` 可以作为用户唯一标识，是否可以替代 `user_id`？

| 方案                         | 是否可行 | 说明                                                         |
| ---------------------------- | -------- | ------------------------------------------------------------ |
| 直接用 `openid` 作为用户主键 | ✅ 可行   | 若你项目只对接微信小程序，openid 是稳定主键                  |
| 使用 `user_id`（数字ID）     | ✅ 推荐   | 更通用、便于关联数据库中的其他表（如订单、评论、积分等）     |
| 两者都存                     | ✅ 最推荐 | 保留 `user_id`（主键自增ID），记录 `openid`（平台标识），便于扩展和对接其他渠道（如 Web、App、小程序等） |

 ✅ 使用方式推荐

| 场景             | 建议使用                                                     |
| ---------------- | ------------------------------------------------------------ |
| 用户注册、登录时 | 使用 `openid` 查找/创建用户                                  |
| 数据库关联、查询 | 使用 `user_id`                                               |
| 接口认证         | 生成 Token 时包含 `user_id` 或 `openid`，推荐以 `user_id` 为主 |

### 小程序购买流程
1. 用户下单 → 创建 `order` 记录
2. 创建 `stock_operation` 记录（出库操作）
3. 创建 `stock_operation_item` 记录（商品明细，关联订单ID）
4. 更新商品库存

```
CheckoutOrder()
├── 数据校验和准备阶段
└── processCheckoutTransaction() // 事务处理
    ├── 创建订单
    ├── 记录订单日志
    ├── 创建库存操作记录
    ├── 处理库存出库
    └── 删除购物车
```

- order 表专注于订单业务逻辑（支付状态、收货信息等）
- stock_operation 表专注于库存操作记录日志(入库、出库、退货等)
- order_log：订单业务操作日志(创建、取消、删除、支付等)
- stock_operation_item 统一记录所有商品明细，避免数据重复

#### 2. 管理员后台操作流程
1. 管理员操作 → 创建 `stock_operation` 记录
2. 创建 `stock_operation_item` 记录（商品明细）
3. 更新商品库存

### 库存管理功能

#### 1. 小程序用户购买自动出库
- 当用户在小程序中下单时，系统会自动：
  - 检查商品库存是否充足
  - 创建订单记录
  - 自动减少商品库存
  - 记录库存出库日志

#### 2. 管理员后台库存操作
- **批量入库操作** (`POST /admin/stock/batch/inbound`)
  - 管理员可以一次性对多个商品进行入库
  - 记录操作人和操作人ID
  - 记录入库日志
  
- **批量出库操作** (`POST /admin/stock/batch/outbound`)
  - 管理员可以一次性对多个商品进行出库
  - 记录用户名称、用户ID、用户账号、购买时间
  - 检查库存是否充足
  - 记录出库日志

#### 3. 库存操作查询
- **获取库存操作列表** (`GET /admin/stock/operations`)
  - 支持分页查询
  - 显示所有库存操作记录
  
- **获取库存操作详情** (`GET /admin/stock/operation/:id`)
  - 显示操作主表信息
  - 显示操作子表商品明细
  - 记录所有库存操作历史（兼容旧版本）

1. **小程序购买流程**：

   - 用户选择商品并下单
   - 系统自动检查库存
   - 创建订单并减少库存
   - 记录出库日志

2. **管理员库存管理**：

   - 通过后台管理界面进行入库、出库、退货操作
   - 所有操作都会记录详细的日志
   - 支持库存日志查询和追溯

3. **库存安全**：

   - 所有库存操作都有库存检查
   - 出库时会验证库存是否充足
   - 完整的操作日志记录

   

## 测试建议

### 1. 功能测试

- 正常下单流程
- 库存不足时的处理
- 并发下单场景

### 2. 异常测试

- 数据库连接异常
- 网络超时
- 死锁场景

### 3. 性能测试

- 高并发下单
- 事务执行时间
- 数据库连接池使用情况

## 后续优化建议

### 1. 库存锁定机制

- 考虑实现库存预占机制
- 减少事务执行时间

### 2. 异步处理

- 非关键操作可以异步处理
- 提高响应速度

### 3. 监控告警

- 添加事务执行时间监控
- 设置异常告警机制

## API 接口说明

### 用户管理接口

#### 用户登录

```bash
curl --location 'http://127.0.0.1:8009/api/user/login' \
--header 'Content-Type: application/json' \
--data '{
    "code": "wx_login_code_from_miniprogram",
    "nickname": "用户昵称",
    "avatar": "头像URL",
    "latitude": 39.9042,
    "longitude": 116.4074
}'
```

**说明：**
- `code`: 微信登录code（必填）
- `nickname`: 微信昵称（可选）
- `avatar`: 头像URL（可选）
- `latitude`: 纬度（可选，用于确定最近店铺）
- `longitude`: 经度（可选，用于确定最近店铺）

**店铺分配逻辑：**
- 如果提供了位置信息，系统会根据经纬度计算最近的店铺
- 如果距离涞水店更近且距离小于50公里，分配涞水店
- 否则默认分配燕郊店
- 如果没有提供位置信息，默认分配燕郊店

#### 更新用户信息

```bash
curl --location 'http://127.0.0.1:8009/api/user/update/info' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "nickname": "用户昵称",
    "avatar": "头像URL"
}'
```

#### 绑定手机号

```bash
curl --location 'http://127.0.0.1:8009/api/user/bind-mobile' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "mobile": "13800138000"
}'
```

### 商品管理接口

#### 获取商品列表

```bash
curl --location 'http://127.0.0.1:8009/api/product/list' \
--header 'Authorization: Bearer your_jwt_token'
```

**说明：**
- 需要JWT token认证
- 系统会根据用户所属店铺返回对应的商品列表
- 每个用户只能看到自己店铺的商品

### 地址管理接口

#### 获取地址列表

```bash
curl --location 'http://127.0.0.1:8009/api/address/list' \
--header 'Authorization: Bearer your_jwt_token'
```

#### 创建地址

```bash
curl --location 'http://127.0.0.1:8009/api/address/create' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "data": {
        "recipient_name": "收货人姓名",
        "recipient_phone": "13800138000",
        "province": "广东省",
        "city": "深圳市",
        "district": "南山区",
        "detail": "详细地址",
        "is_default": true
    }
}'
```

#### 设置默认地址

```bash
curl --location 'http://127.0.0.1:8009/api/address/set_default/1' \
--header 'Authorization: Bearer your_jwt_token'
```

#### 更新地址

```bash
curl --location 'http://127.0.0.1:8009/api/address/update' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "data": {
        "address_id": 1,
        "recipient_name": "收货人姓名",
        "recipient_phone": "13800138000",
        "province": "广东省",
        "city": "深圳市",
        "district": "南山区",
        "detail": "详细地址",
        "is_default": false
    }
}'
```

#### 删除地址

```bash
curl --location --request DELETE 'http://127.0.0.1:8009/api/address/delete/1' \
--header 'Authorization: Bearer your_jwt_token'
```

### 购物车管理接口

#### 获取购物车列表

```bash
curl --location 'http://127.0.0.1:8009/api/cart/list' \
--header 'Authorization: Bearer your_jwt_token'
```

#### 添加到购物车

```bash
curl --location 'http://127.0.0.1:8009/api/cart/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "product_id": 1,
    "quantity": 2
}'
```

#### 更新购物车商品

```bash
curl --location 'http://127.0.0.1:8009/api/cart/update' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "cart_id": 1,
    "quantity": 3
}'
```

#### 删除购物车商品

```bash
curl --location --request DELETE 'http://127.0.0.1:8009/api/cart/delete/1' \
--header 'Authorization: Bearer your_jwt_token'
```

### 订单管理接口

#### 获取订单列表

```bash
curl --location 'http://127.0.0.1:8009/api/order/list' \
--header 'Authorization: Bearer your_jwt_token'
```

#### 获取订单详情

```bash
curl --location 'http://127.0.0.1:8009/api/order/detail?order_id=1' \
--header 'Authorization: Bearer your_jwt_token'
```

#### 创建订单

```bash
curl --location 'http://127.0.0.1:8009/api/order/checkout' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "address_id": 1,
    "remark": "订单备注"
}'
```

#### 取消订单

```bash
curl --location 'http://127.0.0.1:8009/api/order/cancel' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "order_id": 1
}'
```

#### 删除订单

```bash
curl --location --request DELETE 'http://127.0.0.1:8009/api/order/delete/1' \
--header 'Authorization: Bearer your_jwt_token'
```

### 支付管理接口

#### 获取支付数据

```bash
curl --location 'http://127.0.0.1:8009/api/pay/data' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer your_jwt_token' \
--data '{
    "order_id": 1
}'
```

#### 支付回调

```bash
curl --location 'http://127.0.0.1:8009/api/pay/callback' \
--header 'Content-Type: application/json' \
--data '{
    "callback_data": "支付回调数据"
}'
```

## Admin 接口说明

### 用户管理接口

#### 后台添加用户

```bash
curl --location 'http://127.0.0.1:8009/admin/user/add' \
--header 'Content-Type: application/json' \
--data '{
    "admin_display_name": "孙阳",
    "mobile_phone": "13800138001",
    "shop_id": 1,
    "remark": "塑雅雕塑"
}'
{"code":0,"data":{"id":2,"openid":"","nickname":"","avatar":"","mobile_phone":"13800138001","source":2,"is_enable":1,"admin_display_name":"孙阳","wechat_display_name":"","has_wechat_bind":0,"shop_id":1,"created_at":"2025-09-14T14:28:51.445+08:00","updated_at":"2025-09-14T14:28:51.445+08:00"},"message":"添加用户成功"}
```

**说明：**
- `shop_id`: 店铺ID（可选，不传则默认分配燕郊店）
  - `1`: 燕郊店
  - `2`: 涞水店

#### 后台编辑用户

```bash
curl -X PUT 'http://127.0.0.1:8009/admin/user/edit' \
--header 'Content-Type: application/json' \
--data '{
    "id": 2,
    "admin_display_name": "孙阳（更新）",
    "mobile_phone": "13800138002",
    "is_enable": 1
}'
{"code":0,"data":null,"message":"更新用户成功"}
```

#### 后台获取用户列表

```bash
curl --location 'http://127.0.0.1:8009/admin/user/list?page=1&page_size=10'
{"code":0,"data":{"users":[{"id":2,"openid":"","nickname":"","avatar":"","mobile_phone":"13800138001","source":2,"is_enable":1,"admin_display_name":"孙阳","wechat_display_name":"","has_wechat_bind":0,"created_at":"2025-09-14T14:28:51.445+08:00","updated_at":"2025-09-14T14:28:51.445+08:00"}],"total":1},"message":"获取用户列表成功"}
```

#### 后台获取用户详情

```bash
curl --location 'http://127.0.0.1:8009/admin/user/2'
{"code":0,"data":{"id":2,"openid":"","nickname":"","avatar":"","mobile_phone":"13800138001","source":2,"is_enable":1,"admin_display_name":"孙阳","wechat_display_name":"","has_wechat_bind":0,"created_at":"2025-09-14T14:28:51.445+08:00","updated_at":"2025-09-14T14:28:51.445+08:00"},"message":"获取用户详情成功"}
```

#### 后台删除用户

```bash
curl --location --request DELETE 'http://127.0.0.1:8009/admin/user/del/2'
{"code":0,"data":null,"message":"删除用户成功"}
```

### 地址管理接口

#### 后台获取地址列表

```bash
curl --location 'http://127.0.0.1:8009/admin/address/list?user_id=123&user_name=张三&page=1&page_size=10'
{"code":0,"data":{"list":[{"address_id":1,"user_id":123,"user_name":"张三","recipient_name":"李四","recipient_phone":"13800138000","province":"广东省","city":"深圳市","district":"南山区","detail":"科技园路1号","is_default":true,"created_at":"2024-01-15T10:30:00Z","updated_at":"2024-01-15T10:30:00Z"}],"total":1,"page":1,"page_size":10},"message":"获取地址列表成功"}
```

#### 后台新增地址

```bash
curl --location 'http://127.0.0.1:8009/admin/address/add' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 123,
    "recipient_name": "李四",
    "recipient_phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路1号",
    "is_default": true
}'
{"code":0,"message":"创建地址成功"}
```

#### 后台编辑地址

```bash
curl --location --request PUT 'http://127.0.0.1:8009/admin/address/edit' \
--header 'Content-Type: application/json' \
--data '{
    "id": 1,
    "user_id": 123,
    "recipient_name": "李四（更新）",
    "recipient_phone": "13800138001",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路2号",
    "is_default": false
}'
{"code":0,"message":"更新地址成功"}
```

#### 后台删除地址

```bash
curl --location --request DELETE 'http://127.0.0.1:8009/admin/address/del/1'
{"code":0,"message":"删除地址成功"}
```

**响应字段说明**:

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int64 | 用户ID |
| openid | string | 微信OpenID（后台添加的用户为空） |
| nickname | string | 微信昵称（后台添加的用户为空） |
| avatar | string | 头像（后台添加的用户为空） |
| mobile_phone | string | 手机号 |
| source | int8 | 用户来源(1:小程序,2:后台添加,3:混合) |
| is_enable | int8 | 是否启用(1:启用,0:禁用) |
| admin_display_name | string | 后台管理系统显示的客户名称 |
| wechat_display_name | string | 微信小程序显示的客户名称 |
| has_wechat_bind | int8 | 是否已绑定微信(1:是,0:否) |
| created_at | string | 创建时间 |
| updated_at | string | 更新时间 |

### 商品管理接口

#### 注意事项

1. **金额处理**: 所有金额字段在JSON中显示为元，但系统内部存储为分
2. **必填字段**: 新增和编辑商品时，`name`、`category_id`、`image`为必填字段
3. **图片上传**: 图片上传接口返回的是完整的URL地址
4. **分页参数**: 页码从1开始，每页大小默认为10
5. **商品状态**: `is_on_shelf`字段控制商品是否上架，1表示上架，0表示下架
6. **店铺管理**: 
   - 获取商品列表支持 `shop_id` 参数进行店铺筛选
   - 新增和编辑商品时，`shop_id` 字段可选，如果未提供则从JWT token中获取
   - 超级管理员(root)可以操作所有店铺的商品，其他管理员只能操作自己店铺的商品
   - **权限验证机制**：
     - 前端可以传递 `shop_id` 参数，便于显示当前操作的店铺
     - 后端会验证前端传递的 `shop_id` 是否与管理员权限匹配
     - 如果前端未传递 `shop_id`，则自动使用JWT token中的店铺ID
     - 普通管理员(lizengchun/zhangweiyang)只能操作自己店铺的数据
     - 超级管理员(root)可以操作任意店铺的数据

6. **字段对比**

| 操作     | 商品名称 | 商品分类 | 商品图片 | 售价   | 规格   | 单位   | 备注   | 状态   | 成本相关 |
| :------- | :------- | :------- | :------- | :----- | :----- | :----- | :----- | :----- | :------- |
|          |          |          |          |        |        |        |        |        |          |
| 添加商品 | ✅ 必填   | ✅ 必填   | ✅ 必填   | ✅ 必填 | ✅ 可选 | ✅ 必填 | ✅ 可选 | ✅ 必填 | ❌ 去掉   |
| 编辑商品 | ✅ 必填   | ❌ 去掉   | ✅ 必填   | ✅ 必填 | ✅ 可选 | ❌ 去掉 | ✅ 可选 | ✅ 必填 |          |

---

#### 错误码说明

- `0`: 操作成功
- `-1`: 操作失败，具体错误信息在message字段中

常见错误信息：

- "参数错误: ..." - 请求参数格式错误
- "商品ID格式错误" - 路径参数ID格式不正确
- "添加商品失败: ..." - 数据库操作失败
- "编辑商品失败: ..." - 更新操作失败
- "删除失败" - 删除操作失败
- "获取商品信息失败: ..." - 查询操作失败

#### 接口示例

##### 获取商品列表

```
# 超级管理员(root) - 可以查看所有店铺的商品
curl "http://127.0.0.1:8009/admin/product/list?page=1&page_size=10" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 超级管理员(root) - 可以查看指定店铺的商品
curl "http://127.0.0.1:8009/admin/product/list?page=1&page_size=10&shop_id=2" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 普通管理员(lizengchun) - 只能查看自己店铺的商品
curl "http://127.0.0.1:8009/admin/product/list?page=1&page_size=10" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"

# 普通管理员(lizengchun) - 尝试查看其他店铺商品会返回403错误
curl "http://127.0.0.1:8009/admin/product/list?page=1&page_size=10&shop_id=2" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"
# 返回: {"code":-1,"message":"无权限查看该店铺的商品"}
```



##### 新增商品

```bash
# 普通管理员(lizengchun) - 新增商品到自己的店铺
curl -X POST "http://127.0.0.1:8009/admin/product/add" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "贸彩1K白",
    "category_id": 1,
    "seller_price": 120,
    "image": "http://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/07GE2k1DJWhah4QA_RlY91685434479136.jpg",
    "unit": "桶",
    "is_on_shelf": 1,
    "remark": "",
    "shop_id": 1
  }'

# 普通管理员(lizengchun) - 不传递shop_id，自动使用JWT中的店铺ID
curl -X POST "http://127.0.0.1:8009/admin/product/add" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "贸彩1K白",
    "category_id": 1,
    "seller_price": 120,
    "image": "http://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/07GE2k1DJWhah4QA_RlY91685434479136.jpg",
    "unit": "桶",
    "is_on_shelf": 1,
    "remark": ""
  }'

# 普通管理员(lizengchun) - 尝试添加商品到其他店铺会返回403错误
curl -X POST "http://127.0.0.1:8009/admin/product/add" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "贸彩1K白",
    "category_id": 1,
    "seller_price": 120,
    "image": "http://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/07GE2k1DJWhah4QA_RlY91685434479136.jpg",
    "unit": "桶",
    "is_on_shelf": 1,
    "remark": "",
    "shop_id": 2
  }'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

##### 编辑商品

```bash
# 编辑商品（店铺ID从JWT token中获取）
curl -X PUT "http://127.0.0.1:8009/admin/product/edit/4" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "贸彩1K白",
    "seller_price": 120,
    "image": "http://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/07GE2k1DJWhah4QA_RlY91685434479136.jpg",
    "is_on_shelf": 1,
    "remark": "",
    "shop_id": 1
  }'
```

##### 删除商品
```http
curl --location --request DELETE 'http://127.0.0.1:8009/admin/product/del/1'
```

##### 获取商品分类

**说明：**
- 支持 `shop_id` 参数进行店铺筛选
- 超级管理员(root)可以查看所有店铺的分类，普通管理员只能查看自己店铺的分类
- 如果未传递 `shop_id` 参数，则自动使用JWT token中的店铺ID

```bash
# 超级管理员(root) - 查看所有分类
curl "http://127.0.0.1:8009/admin/product/categories" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 超级管理员(root) - 查看指定店铺的分类
curl "http://127.0.0.1:8009/admin/product/categories?shop_id=2" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 普通管理员(lizengchun) - 只能查看自己店铺的分类
curl "http://127.0.0.1:8009/admin/product/categories" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"

# 普通管理员(lizengchun) - 尝试查看其他店铺分类会返回403错误
curl "http://127.0.0.1:8009/admin/product/categories?shop_id=2" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**响应示例**
```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "1K-漆",
      "sort_order": 100,
      "shop_id": 1
    },
    {
      "id": 2,
      "name": "2K-漆", 
      "sort_order": 99,
      "shop_id": 1
    }
  ]
}
```

##### 新增商品分类

**说明：**
- 需要传递 `shop_id` 参数指定店铺
- 超级管理员(root)可以创建任意店铺的分类，普通管理员只能创建自己店铺的分类
- 系统会验证 `shop_id` 与管理员权限是否匹配

```bash
# 普通管理员(lizengchun) - 新增分类到自己的店铺
curl -X POST "http://127.0.0.1:8009/admin/product/category/add" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试分类1",
    "sort_order": 96,
    "shop_id": 1
  }'

# 超级管理员(root) - 新增分类到指定店铺
curl -X POST "http://127.0.0.1:8009/admin/product/category/add" \
  -H "Authorization: Bearer ROOT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "涞水店专用分类",
    "sort_order": 95,
    "shop_id": 2
  }'

# 普通管理员尝试创建其他店铺分类会返回403错误
curl -X POST "http://127.0.0.1:8009/admin/product/category/add" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "其他店铺分类",
    "sort_order": 94,
    "shop_id": 2
  }'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
{"code":0,"message":"添加分类成功"}
```

##### 编辑商品分类

**说明：**
- 需要传递 `shop_id` 参数指定店铺
- 超级管理员(root)可以编辑任意店铺的分类，普通管理员只能编辑自己店铺的分类
- 系统会验证 `shop_id` 与管理员权限是否匹配

```bash
# 普通管理员(lizengchun) - 编辑自己店铺的分类
curl -X PUT "http://127.0.0.1:8009/admin/product/category/edit/4" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试分类4",
    "sort_order": 96,
    "shop_id": 1
  }'

# 超级管理员(root) - 编辑指定店铺的分类
curl -X PUT "http://127.0.0.1:8009/admin/product/category/edit/5" \
  -H "Authorization: Bearer ROOT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "涞水店分类更新",
    "sort_order": 95,
    "shop_id": 2
  }'

# 普通管理员尝试编辑其他店铺分类会返回403错误
curl -X PUT "http://127.0.0.1:8009/admin/product/category/edit/5" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "其他店铺分类",
    "sort_order": 94,
    "shop_id": 2
  }'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**响应示例**
```json
{"code": 0, "message": "编辑分类成功"}
```

##### 删除商品分类

**说明：**
- 无需传递 `shop_id` 参数，系统会从JWT token中获取管理员权限
- 系统会先查询分类信息，然后验证管理员是否有权限删除该分类
- 超级管理员(root)可以删除任意店铺的分类，普通管理员只能删除自己店铺的分类

```bash
# 普通管理员(lizengchun) - 删除自己店铺的分类
curl -X DELETE "http://127.0.0.1:8009/admin/product/category/del/4" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"

# 超级管理员(root) - 删除任意店铺的分类
curl -X DELETE "http://127.0.0.1:8009/admin/product/category/del/5" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 普通管理员尝试删除其他店铺分类会返回403错误
curl -X DELETE "http://127.0.0.1:8009/admin/product/category/del/5" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"
# 返回: {"code":-1,"message":"无权限删除该分类"}
```

**响应示例**
```json
{"code": 0, "message": "删除分类成功"}
```



### 库存管理接口

#### 1. 批量入库操作

**说明：**
- 批量入库接口的单个item对象已简化，只保留核心字段
- 前端传递：`product_id`、`quantity`、`product_cost`（进价）、`total_price`（单个商品总价）、`remark`
- `ProductName`、`Specification`、`Unit` 从 Product 表里查询获取
- 入库时只更新 Product 表的 `product_cost` 字段（进价）
- Product 表的 `shipping_cost` 字段在初始化时设置，且不变
- 总金额由前端计算并传递
- **必须指定店铺ID**，管理员手动选择哪个店铺进行入库

```
➜  ~ curl --location 'http://127.0.0.1:8009/admin/stock/batch/inbound' \
--header 'Content-Type: application/json' \
--data '{
"items": [
{
"product_id": 2,
"quantity": 20,
"product_cost": 66,
"total_price": 1320,
"remark": ""
}
],
"total_amount": 1320,
"operator": "张三",
"operator_id": 1001,
"shop_id": 1,
"supplier": "李彦鹏",
"remark": "0901入库"
}'
{"code":0,"message":"批量入库成功"}%
```

**新增字段说明：**
- `shop_id`: 店铺ID（必填）
  - `1`: 燕郊店
  - `2`: 涞水店

#### 2. 批量出库操作
```http
POST /admin/stock/batch/outbound
Content-Type: application/json

{
    "items": [
        {
            "product_id": 3,
            "quantity": 2,
            "unit_price": 85,
            "total_price": 170,
            "remark": "调鼻尖"
        }
    ],
    "total_amount": 170,
    "user_name": "李四",
    "user_id": 1002,
    "operator": "张三",
    "operator_id": 1001,
    "shop_id": 1,
    "remark": "后台操作"
}
```

**新增字段说明：**
- `shop_id`: 店铺ID（必填）
  - `1`: 燕郊店
  - `2`: 涞水店

**说明：**
- 批量出库接口的单个item对象已简化，只保留核心字段
- 前端传递：`product_id`、`quantity`、`unit_price`（可选，不传则使用商品售价）、`total_price`、`remark`
- `ProductName`、`Specification`、`Unit` 从 Product 表里查询获取，减少数据传输压力
- 总金额由前端计算并传递

**响应示例：**
```json
{
  "code": 0,
  "message": "批量出库成功"
}
```

#### 3. 获取库存操作列表
```http
GET /admin/stock/operations?page=1&page_size=10&types=2
```

**查询参数：**
- `page`: 页码，默认为1
- `page_size`: 每页大小，默认为10
- `types`: 操作类型（可选），1-入库，2-出库，3-退货

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "operation_no": "STOCK202508211002009854",
        "types": 2,
        "operator": "张三",
        "operator_id": 1001,
        "operator_type": 2,
        "user_name": "李四",
        "user_id": 1002,
        "user_account": "13671210659",
        "remark": "后台操作",
        "total_amount": 170.00,
        "total_quantity": 2,
        "items": [
          {
            "id": 1,
            "operation_id": 1,
            "product_id": 3,
            "product_name": "固态灰",
            "specification": "3KG*4",
            "quantity": 2,
            "unit_price": 85.00,
            "total_price": 170.00,
            "before_stock": 100,
            "after_stock": 98,
            "remark": "调鼻尖"
          }
        ]
      }
    ],
    "page": 1,
    "page_size": 10,
    "total": 1
  }
}
```

#### 4. 获取库存操作详情
```http
GET /admin/stock/operation/123
```

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "operation_id": 1,
        "product_id": 3,
        "product_name": "固态灰",
        "specification": "3KG*4",
        "quantity": 2,
        "unit_price": 85.00,
        "total_price": 170.00,
        "before_stock": 100,
        "after_stock": 98,
        "remark": "调鼻尖"
      }
    ],
    "operation": {
      "id": 1,
      "operation_no": "STOCK202508211002009854",
      "types": 2,
      "operator": "张三",
      "operator_id": 1001,
      "operator_type": 2,
      "user_name": "李四",
      "user_id": 1002,
      "user_account": "13671210659",
      "remark": "后台操作",
      "total_amount": 170.00,
      "total_quantity": 2,
      "items": null
    }
  }
}
```

#### 5. 更新出库单支付状态
```http
POST /admin/stock/set/payment-status
Content-Type: application/json

{
  "operation_id": 123,
  "payment_finish_status": 3,
  "operator": "管理员",
  "operator_id": 1001
}
```

**curl 命令示例：**

**设置为已支付：**
```bash
curl --location 'http://127.0.0.1:8009/admin/stock/set/payment-status' \
--header 'Content-Type: application/json' \
--data '{
  "operation_id": 123,
  "payment_finish_status": 3,
  "operator": "管理员",
  "operator_id": 1001
}'
```

**响应示例：**
```json
{
  "code": 0,
  "message": "更新支付状态成功"
}
```

**请求参数说明：**
- `operation_id`: 出库单ID（必填）
- `payment_finish_status`: 支付完成状态（必填）
  - `1`: 未支付
  - `3`: 已支付
- `operator`: 操作人姓名（必填）
- `operator_id`: 操作人ID（必填）

**业务说明：**
- 新建出库单时默认状态为未支付（1）
- 客户私下转账后，管理员调用此接口设置为已支付（3）
- 设置为已支付时，系统会自动记录支付完成时间
- 只能更新出库单的支付状态，不能更新入库单

#### 6. 获取供货商列表
```http
GET /admin/stock/suppliers
```

**curl 命令示例：**
```bash
curl --location 'http://127.0.0.1:8009/admin/stock/suppliers'
```

**响应示例：**
```json
{
  "code": 0,
  "message": "获取供货商列表成功",
  "data": [
    {
      "id": 1,
      "name": "华润涂料有限公司",
      "area": "广东省佛山市",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "name": "立邦涂料（中国）有限公司",
      "area": "上海市",
      "created_at": "2024-01-16T14:20:00Z",
      "updated_at": "2024-01-16T14:20:00Z"
    }
  ]
}
```

**响应字段说明：**
- `id`: 供货商ID
- `name`: 供货商名称
- `area`: 供货商所在地区
- `created_at`: 创建时间
- `updated_at`: 更新时间

#### 字段说明

**批量入库请求字段：**
- `items`: 入库商品列表
  - `product_id`: 商品ID（必填）
  - `quantity`: 入库数量（必填）
  - `cost`: 成本价（必填，单位：分）
  - `shipping_cost`: 运费成本（必填，单位：分）
  - `product_cost`: 货物成本（必填，单位：分）
  - `remark`: 备注（可选）
  - `product_name`: 商品全名（自动补齐，前端可传空字符串）
  - `specification`: 规格（自动补齐，前端可传空字符串）
  - `unit`: 单位（自动补齐，前端可传空字符串）
  - `total_amount`: 总金额（自动计算，前端可传0）
- `total_amount`: 总金额（前端计算，单位：分）
- `operator`: 操作人姓名（必填）
- `operator_id`: 操作人ID（必填）
- `supplier`: 供货商（可选）
- `remark`: 操作备注（可选）

**批量出库请求字段：**
- `items`: 出库商品列表
  - `product_id`: 商品ID（必填）
  - `quantity`: 出库数量（必填）
  - `unit_price`: 单价（可选，不传则使用商品售价，单位：分）
  - `total_price`: 总金额（必填，单位：分）
  - `remark`: 备注（可选）
  - `product_name`: 商品全名（从商品表获取，前端无需传递）
  - `specification`: 规格（从商品表获取，前端无需传递）
  - `unit`: 单位（从商品表获取，前端无需传递）
- `total_amount`: 总金额（前端计算，单位：分）
- `user_name`: 用户名称（必填）
- `user_id`: 用户ID（必填）
- `operator`: 操作人姓名（必填）
- `operator_id`: 操作人ID（必填）
- `remark`: 操作备注（可选）

**操作类型说明：**
- `types`: 1-入库, 2-出库, 3-退货
- `outbound_type`: 1-小程序购买, 2-admin后台操作（仅出库时有效）
- `operator_type`: 1-用户, 2-系统, 3-管理员

**查询参数说明：**
- `types` 查询参数用于过滤特定类型的库存操作：
  - 不传 `types` 参数：查询所有类型的操作
  - `types=1`：只查询入库操作
  - `types=2`：只查询出库操作
  - `types=3`：只查询退货操作

**使用示例：**
```bash
# 查询所有操作
curl 'http://192.168.99.172:8009/admin/stock/operations'

# 查询入库操作
curl 'http://192.168.99.172:8009/admin/stock/operations?types=1'

# 查询出库操作
curl 'http://192.168.99.172:8009/admin/stock/operations?types=2'

# 查询退货操作
curl 'http://192.168.99.172:8009/admin/stock/operations?types=3'

# 结合分页查询出库操作
curl 'http://192.168.99.172:8009/admin/stock/operations?types=2&page=1&page_size=10'
```

**说明：**
- 入库时：后端会自动补齐商品信息（`product_name`, `specification`, `unit`），前端在items中传入这些字段时可以使用空字符串，后端会自动填充
- 出库时：后端会自动从商品表获取商品信息（`product_name`, `specification`, `unit`），前端无需传递这些字段，减少数据传输压力
- 入库时：如果没有提供 `unit_price`，会使用商品的 `seller_price`；如果新成本价更低，会自动更新商品成本价并记录变更
- 出库时：如果没有提供 `unit_price`，会使用商品的 `seller_price`
- 时间字段由后端自动记录，无需前端传入



---

## 业务逻辑说明

### 1. 默认地址处理

- 当设置某个地址为默认地址时，系统会自动取消该用户的其他默认地址
- 每个用户只能有一个默认地址

### 2. 权限控制

- 管理员可以为任意用户创建、编辑、删除地址
- 删除地址时需要提供用户ID进行权限验证

### 3. 数据安全

- 删除地址采用软删除方式，不会真正删除数据库记录
- 所有操作都会记录操作日志

### 4. 搜索功能

- 支持通过用户ID精确搜索
- 支持通过用户名称模糊搜索
- 两个搜索条件可以同时使用

### 5. 数据验证

- 收货人姓名、电话、省市区、详细地址为必填字段
- 电话号码格式建议进行前端验证
- 地址信息建议进行合理性验证
- 

## 数据库表结构

### 商品表

#### product 商品表
```sql
CREATE TABLE product (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL COMMENT '商品全名',
  category_id BIGINT NOT NULL COMMENT '分类ID',
  image VARCHAR(500) COMMENT '商品图片',
  seller_price BIGINT NOT NULL DEFAULT 0 COMMENT '单价(分)',
  cost BIGINT DEFAULT 0 COMMENT '成本(分)',
  shipping_cost BIGINT DEFAULT 0 COMMENT '运费(分)',
  product_cost BIGINT DEFAULT 0 COMMENT '货物成本(分)',
  specification VARCHAR(200) DEFAULT '' COMMENT '规格',
  unit VARCHAR(50) DEFAULT '' COMMENT '单位',
  remark VARCHAR(500) COMMENT '备注',
  stock INT NOT NULL DEFAULT 0 COMMENT '库存',
  is_on_shelf TINYINT NOT NULL DEFAULT 1 COMMENT '是否上架(1:上架,0:下架)',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
);
```

### 库存操作

#### stock_operation 库存操作主表

```sql
CREATE TABLE stock_operation (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  operation_no VARCHAR(64) NOT NULL COMMENT '操作单号',
  type TINYINT NOT NULL COMMENT '操作类型(1:入库,2:出库,3:退货)',
  operator VARCHAR(255) NOT NULL COMMENT '操作人',
  operator_id BIGINT NOT NULL COMMENT '操作人ID',
  operator_type TINYINT NOT NULL COMMENT '操作人类型(1:用户,2:系统,3:管理员)',
  user_name VARCHAR(255) COMMENT '用户名称(出库时)',
  user_id BIGINT COMMENT '用户ID(出库时)',
  user_account VARCHAR(255) COMMENT '用户账号(出库时)',
  purchase_time TIMESTAMP COMMENT '购买时间(出库时)',
  remark VARCHAR(255) COMMENT '备注',
  total_amount BIGINT NOT NULL DEFAULT 0 COMMENT '总金额(分)',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);
```

#### stock_operation_item 库存操作子表
```sql
CREATE TABLE stock_operation_item (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  operation_id BIGINT NOT NULL COMMENT '操作主表ID',
  product_id BIGINT NOT NULL COMMENT '商品ID',
  product_name VARCHAR(255) NOT NULL COMMENT '商品全名',
  specification VARCHAR(200) DEFAULT '' COMMENT '规格',
  quantity INT NOT NULL COMMENT '操作数量',
  unit_price BIGINT NOT NULL DEFAULT 0 COMMENT '单价(分)',
  total_price BIGINT NOT NULL DEFAULT 0 COMMENT '总价(分)',
  before_stock INT NOT NULL COMMENT '操作前库存',
  after_stock INT NOT NULL COMMENT '操作后库存',
  cost BIGINT NOT NULL DEFAULT 0 COMMENT '成本价(暂不用) 单位:分',
  shipping_cost BIGINT NOT NULL DEFAULT 0 COMMENT '运费(暂不用) 单位:分',
  product_cost BIGINT NOT NULL DEFAULT 0 COMMENT '货物成本(暂不用) 单位:分',
  remark VARCHAR(500) DEFAULT '' COMMENT '备注',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);
```

#### inbound_cost_change 入库成本变更记录表
```sql
CREATE TABLE inbound_cost_change (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键id',
  operation_id BIGINT NOT NULL COMMENT '入库操作ID',
  product_id BIGINT NOT NULL COMMENT '商品ID',
  product_name VARCHAR(255) NOT NULL COMMENT '商品名称',
  old_cost BIGINT NOT NULL DEFAULT 0 COMMENT '原成本价(分)',
  new_cost BIGINT NOT NULL DEFAULT 0 COMMENT '新成本价(分)',
  change_reason VARCHAR(500) DEFAULT '' COMMENT '变更原因',
  operator VARCHAR(100) NOT NULL COMMENT '操作人',
  operator_id BIGINT NOT NULL COMMENT '操作人ID',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  INDEX idx_operation_id (operation_id),
  INDEX idx_product_id (product_id),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='入库成本变更记录表';
```

#### stock_log 库存日志表（兼容旧版本）
```sql
CREATE TABLE stock_log (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  product_id BIGINT NOT NULL COMMENT '商品ID',
  product_name VARCHAR(255) NOT NULL COMMENT '商品名称',
  type TINYINT NOT NULL COMMENT '操作类型(1:入库,2:出库,3:退货)',
  quantity INT NOT NULL COMMENT '操作数量',
  before_stock INT NOT NULL COMMENT '操作前库存',
  after_stock INT NOT NULL COMMENT '操作后库存',
  order_no VARCHAR(64) COMMENT '关联订单号(出库/退货时)',
  remark VARCHAR(255) COMMENT '备注',
  operator VARCHAR(255) NOT NULL COMMENT '操作人',
  operator_id BIGINT COMMENT '操作人ID',
  operator_type TINYINT NOT NULL COMMENT '操作人类型(1:用户,2:系统,3:管理员)',
  buyer_name VARCHAR(255) COMMENT '购买者名称(出库时)',
  buyer_account VARCHAR(255) COMMENT '购买者账号(出库时)',
  purchase_time TIMESTAMP COMMENT '购买时间(出库时)',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);
```

