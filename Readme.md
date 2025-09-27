# 油漆销售系统

## 业务说明

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

#### (技术)数据库初始化

首次使用前，需要在数据库中插入店铺数据：

### 后台管理认证

#### 权限控制

- **超级管理员 (root)**: 可以操作所有店铺的数据

- **普通管理员(lizengchun、zhangweiyang)**: 只能操作自己所属店铺的数据

  Token 有效期: 2小时

  自动店铺关联: 所有操作自动关联到管理员所属店铺

#### (技术)数据库初始化
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

### 关于账号&登陆

#### 1.小程序登陆

1. 小程序的openid 作为用户在微信的唯一标识 仅第一次登陆，发起支付用。


| 阶段             | 操作                                       | 调用 GetOpenIDByCode？ | 后端是否需要用它？ |
| ---------------- | ------------------------------------------ | ---------------------- | ------------------ |
| 第一次登录小程序 | wx.login() + 后端换取 openid + 绑定 userID | ✅ 是                   | ✅ 是               |
| 后续所有请求     | 带上 token 或 userID                       | ❌ 否                   | ❌ 不需要调用       |
| 发起支付         | 用存下来的 openid 发起支付下单             | ❌ 否                   | ✅ 用 openid        |

2. 我们系统会生成唯一的user_id 与 open_id 进行绑定

#### 2.web后台添加账号

### 关于下单

#### 1. 小程序购买

```
CheckoutOrder()
├── 数据校验和准备阶段
└── processCheckoutTransaction() // 事务处理
    ├── 创建订单 >order表
    ├── 记录订单日志 >order_log表
    ├── 创建库存操作记录 >stock_operation表、stock_operation_log表
    ├── 处理库存出库 > 更新product表
    └── 删除购物车 > 更新cart表
```

- order 表专注于订单业务逻辑（支付状态、收货信息等）
- stock_operation 表专注于库存操作记录日志(入库、出库、退货等)
- order_log：订单业务操作日志(创建、取消、删除、支付等)
- stock_operation_item 统一记录所有商品明细，避免数据重复

#### 2. 管理员后台创建出库单
## TODO后续优化建议

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

**说明：**
- 支持 `shop_id` 参数进行店铺筛选
- 超级管理员(root)可以添加任何店铺的用户，普通管理员只能添加自己店铺的用户
- 如果未传递 `shop_id` 参数，则自动使用JWT token中的店铺ID

```bash
# 超级管理员(root) - 添加用户到指定店铺
curl --location 'http://127.0.0.1:8009/admin/user/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer ROOT_TOKEN' \
--data '{
    "admin_display_name": "孙阳",
    "mobile_phone": "13800138001",
    "shop_id": 1,
    "remark": "塑雅雕塑"
}'

# 普通管理员(lizengchun) - 只能添加自己店铺的用户
curl --location 'http://127.0.0.1:8009/admin/user/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN' \
--data '{
    "admin_display_name": "孙阳",
    "mobile_phone": "13800138001",
    "shop_id": 1,
    "remark": "塑雅雕塑"
}'

# 普通管理员(lizengchun) - 尝试添加其他店铺用户会返回403错误
curl --location 'http://127.0.0.1:8009/admin/user/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN' \
--data '{
    "admin_display_name": "孙阳",
    "mobile_phone": "13800138001",
    "shop_id": 2,
    "remark": "塑雅雕塑"
}'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**请求示例：**
```http
POST /admin/user/add
Content-Type: application/json
Authorization: Bearer TOKEN

{
    "admin_display_name": "孙阳",
    "mobile_phone": "13800138001",
    "shop_id": 1,
    "remark": "塑雅雕塑"
}
```

#### 后台编辑用户

**说明：**
- 支持 `shop_id` 参数进行店铺筛选
- 超级管理员(root)可以编辑任何店铺的用户，普通管理员只能编辑自己店铺的用户
- 如果未传递 `shop_id` 参数，则自动使用JWT token中的店铺ID

```bash
# 超级管理员(root) - 编辑用户到指定店铺
curl -X PUT 'http://127.0.0.1:8009/admin/user/edit' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer ROOT_TOKEN' \
--data '{
    "id": 2,
    "admin_display_name": "孙阳（更新）",
    "mobile_phone": "13800138002",
    "is_enable": 1,
    "shop_id": 1
}'

# 普通管理员(lizengchun) - 只能编辑自己店铺的用户
curl -X PUT 'http://127.0.0.1:8009/admin/user/edit' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN' \
--data '{
    "id": 2,
    "admin_display_name": "孙阳（更新）",
    "mobile_phone": "13800138002",
    "is_enable": 1,
    "shop_id": 1
}'

# 普通管理员(lizengchun) - 尝试编辑其他店铺用户会返回403错误
curl -X PUT 'http://127.0.0.1:8009/admin/user/edit' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN' \
--data '{
    "id": 2,
    "admin_display_name": "孙阳（更新）",
    "mobile_phone": "13800138002",
    "is_enable": 1,
    "shop_id": 2
}'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**请求示例：**
```http
PUT /admin/user/edit
Content-Type: application/json
Authorization: Bearer TOKEN

{
    "id": 2,
    "admin_display_name": "孙阳（更新）",
    "mobile_phone": "13800138002",
    "is_enable": 1,
    "shop_id": 1
}
```

#### 后台获取用户列表

**说明：**
- 支持 `shop_id` 参数进行店铺筛选
- 超级管理员(root)可以查看所有店铺的用户，普通管理员只能查看自己店铺的用户
- 如果未传递 `shop_id` 参数，则自动使用JWT token中的店铺ID

```bash
# 超级管理员(root) - 查看所有用户
curl --location 'http://127.0.0.1:8009/admin/user/list?page=1&page_size=10' \
--header 'Authorization: Bearer ROOT_TOKEN'

# 超级管理员(root) - 查看指定店铺的用户
curl --location 'http://127.0.0.1:8009/admin/user/list?page=1&page_size=10&shop_id=2' \
--header 'Authorization: Bearer ROOT_TOKEN'

# 普通管理员(lizengchun) - 只能查看自己店铺的用户
curl --location 'http://127.0.0.1:8009/admin/user/list?page=1&page_size=10' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'

# 普通管理员(lizengchun) - 尝试查看其他店铺用户会返回403错误
curl --location 'http://127.0.0.1:8009/admin/user/list?page=1&page_size=10&shop_id=2' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**请求示例：**
```http
GET /admin/user/list?page=1&page_size=10&shop_id=1
Authorization: Bearer TOKEN
```

#### 后台获取用户详情

**说明：**
- 无需传递 `shop_id` 参数，但会校验用户所属店铺权限
- 超级管理员(root)可以查看任何用户详情，普通管理员只能查看自己店铺的用户详情

```bash
# 超级管理员(root) - 可以查看任何用户详情
curl --location 'http://127.0.0.1:8009/admin/user/2' \
--header 'Authorization: Bearer ROOT_TOKEN'

# 普通管理员(lizengchun) - 只能查看自己店铺的用户详情
curl --location 'http://127.0.0.1:8009/admin/user/2' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'

# 普通管理员(lizengchun) - 尝试查看其他店铺用户详情会返回403错误
curl --location 'http://127.0.0.1:8009/admin/user/3' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'
# 返回: {"code":-1,"message":"无权限查看该用户信息"}
```

**请求示例：**
```http
GET /admin/user/2
Authorization: Bearer TOKEN
```

#### 后台删除用户

**说明：**
- 无需传递 `shop_id` 参数，但会校验用户所属店铺权限
- 超级管理员(root)可以删除任何用户，普通管理员只能删除自己店铺的用户

```bash
# 超级管理员(root) - 可以删除任何用户
curl --location --request DELETE 'http://127.0.0.1:8009/admin/user/del/2' \
--header 'Authorization: Bearer ROOT_TOKEN'

# 普通管理员(lizengchun) - 只能删除自己店铺的用户
curl --location --request DELETE 'http://127.0.0.1:8009/admin/user/del/2' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'

# 普通管理员(lizengchun) - 尝试删除其他店铺用户会返回403错误
curl --location --request DELETE 'http://127.0.0.1:8009/admin/user/del/3' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'
# 返回: {"code":-1,"message":"无权限删除该用户"}
```

**请求示例：**
```http
DELETE /admin/user/del/2
Authorization: Bearer TOKEN
```

### 地址管理接口

#### 后台获取地址列表

**说明：**
- 支持 `shop_id` 参数进行店铺筛选
- 超级管理员(root)可以查看所有店铺的地址，普通管理员只能查看自己店铺的地址
- 如果未传递 `shop_id` 参数，则自动使用JWT token中的店铺ID

```bash
# 超级管理员(root) - 查看所有地址
curl --location 'http://127.0.0.1:8009/admin/address/list?user_id=123&user_name=张三&page=1&page_size=10' \
--header 'Authorization: Bearer ROOT_TOKEN'

# 超级管理员(root) - 查看指定店铺的地址
curl --location 'http://127.0.0.1:8009/admin/address/list?user_id=123&user_name=张三&page=1&page_size=10&shop_id=2' \
--header 'Authorization: Bearer ROOT_TOKEN'

# 普通管理员(lizengchun) - 只能查看自己店铺的地址
curl --location 'http://127.0.0.1:8009/admin/address/list?user_id=123&user_name=张三&page=1&page_size=10' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'

# 普通管理员(lizengchun) - 尝试查看其他店铺地址会返回403错误
curl --location 'http://127.0.0.1:8009/admin/address/list?user_id=123&user_name=张三&page=1&page_size=10&shop_id=2' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**请求示例：**
```http
GET /admin/address/list?user_id=123&user_name=张三&page=1&page_size=10&shop_id=1
Authorization: Bearer TOKEN
```

#### 后台新增地址

**说明：**
- 支持 `shop_id` 参数进行店铺筛选
- 超级管理员(root)可以添加任何店铺的地址，普通管理员只能添加自己店铺的地址
- 如果未传递 `shop_id` 参数，则自动使用JWT token中的店铺ID

```bash
# 超级管理员(root) - 添加地址到指定店铺
curl --location 'http://127.0.0.1:8009/admin/address/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer ROOT_TOKEN' \
--data '{
    "user_id": 123,
    "shop_id": 1,
    "recipient_name": "李四",
    "recipient_phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路1号",
    "is_default": true
}'

# 普通管理员(lizengchun) - 只能添加自己店铺的地址
curl --location 'http://127.0.0.1:8009/admin/address/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN' \
--data '{
    "user_id": 123,
    "shop_id": 1,
    "recipient_name": "李四",
    "recipient_phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路1号",
    "is_default": true
}'

# 普通管理员(lizengchun) - 尝试添加其他店铺地址会返回403错误
curl --location 'http://127.0.0.1:8009/admin/address/add' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN' \
--data '{
    "user_id": 123,
    "shop_id": 2,
    "recipient_name": "李四",
    "recipient_phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路1号",
    "is_default": true
}'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**请求示例：**
```http
POST /admin/address/add
Content-Type: application/json
Authorization: Bearer TOKEN

{
    "user_id": 123,
    "shop_id": 1,
    "recipient_name": "李四",
    "recipient_phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路1号",
    "is_default": true
}
```

#### 后台编辑地址

**说明：**
- 支持 `shop_id` 参数进行店铺筛选
- 超级管理员(root)可以编辑任何店铺的地址，普通管理员只能编辑自己店铺的地址
- 如果未传递 `shop_id` 参数，则自动使用JWT token中的店铺ID

```bash
# 超级管理员(root) - 编辑地址到指定店铺
curl --location --request PUT 'http://127.0.0.1:8009/admin/address/edit' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer ROOT_TOKEN' \
--data '{
    "id": 1,
    "user_id": 123,
    "shop_id": 1,
    "recipient_name": "李四（更新）",
    "recipient_phone": "13800138001",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路2号",
    "is_default": false
}'

# 普通管理员(lizengchun) - 只能编辑自己店铺的地址
curl --location --request PUT 'http://127.0.0.1:8009/admin/address/edit' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN' \
--data '{
    "id": 1,
    "user_id": 123,
    "shop_id": 1,
    "recipient_name": "李四（更新）",
    "recipient_phone": "13800138001",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路2号",
    "is_default": false
}'

# 普通管理员(lizengchun) - 尝试编辑其他店铺地址会返回403错误
curl --location --request PUT 'http://127.0.0.1:8009/admin/address/edit' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN' \
--data '{
    "id": 1,
    "user_id": 123,
    "shop_id": 2,
    "recipient_name": "李四（更新）",
    "recipient_phone": "13800138001",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路2号",
    "is_default": false
}'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**请求示例：**
```http
PUT /admin/address/edit
Content-Type: application/json
Authorization: Bearer TOKEN

{
    "id": 1,
    "user_id": 123,
    "shop_id": 1,
    "recipient_name": "李四（更新）",
    "recipient_phone": "13800138001",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路2号",
    "is_default": false
}
```

#### 后台删除地址

**说明：**
- 无需传递 `shop_id` 参数，但会校验地址所属店铺权限
- 超级管理员(root)可以删除任何地址，普通管理员只能删除自己店铺的地址

```bash
# 超级管理员(root) - 可以删除任何地址
curl --location --request DELETE 'http://127.0.0.1:8009/admin/address/del/1' \
--header 'Authorization: Bearer ROOT_TOKEN'

# 普通管理员(lizengchun) - 只能删除自己店铺的地址
curl --location --request DELETE 'http://127.0.0.1:8009/admin/address/del/1' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'

# 普通管理员(lizengchun) - 尝试删除其他店铺地址会返回403错误
curl --location --request DELETE 'http://127.0.0.1:8009/admin/address/del/2' \
--header 'Authorization: Bearer LIZENGCHUN_TOKEN'
# 返回: {"code":-1,"message":"无权限删除该地址"}
```

**请求示例：**
```http
DELETE /admin/address/del/1
Authorization: Bearer TOKEN
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
7. **商品名称查询**: 
   - 获取商品列表支持 `name` 参数进行商品名称模糊查询
   - 支持与 `shop_id` 参数组合使用，实现店铺+名称的组合筛选
   - 查询使用 `LIKE '%keyword%'` 进行模糊匹配
8. **成本字段管理**: 
   - 添加商品时支持设置成本相关字段：`cost`（成本价）、`shipping_cost`（运费成本）、`product_cost`（货物成本）
   - 成本价 = 运费成本 + 货物成本
   - 这些字段为可选字段，如果不提供则默认为0
   - 编辑商品时不支持修改成本字段，成本由入库操作自动更新
9. **编辑商品字段管理**: 
   - 编辑商品支持部分字段更新，前端传什么字段就更新什么字段，不传的字段保持不变
   - 支持更新的字段：`seller_price`（售价）、`specification`（规格）、`is_on_shelf`（上架状态）、`remark`（备注）、`stock`（库存）
   - 不支持更新的字段：`name`（商品名称）、`image`（商品图片）、`category_id`（分类ID）、`unit`（单位）、成本相关字段
   - 这种设计避免了不必要的字段更新，提高了接口的灵活性和性能
10. **权限验证机制**：
   - 前端可以传递 `shop_id` 参数，便于显示当前操作的店铺
   - 后端会验证前端传递的 `shop_id` 是否与管理员权限匹配
   - 如果前端未传递 `shop_id`，则自动使用JWT token中的店铺ID
   - 普通管理员(lizengchun/zhangweiyang)只能操作自己店铺的数据
   - 超级管理员(root)可以操作任意店铺的数据

11. **字段对比**

| 操作     | 商品名称 | 商品分类 | 商品图片 | 售价   | 规格   | 单位   | 备注   | 状态   | 库存   | 成本价 | 运费成本 | 货物成本 |
| :------- | :------- | :------- | :------- | :----- | :----- | :----- | :----- | :----- | :----- | :----- | :------- | :------- |
|          |          |          |          |        |        |        |        |        |        |        |          |          |
| 添加商品 | ✅ 必填   | ✅ 必填   | ✅ 必填   | ✅ 必填 | ✅ 可选 | ✅ 必填 | ✅ 可选 | ✅ 必填 | ❌ 固定0 | ✅ 可选 | ✅ 可选   | ✅ 可选   |
| 编辑商品 | ❌ 不支持 | ❌ 不支持 | ❌ 不支持 | ✅ 可选 | ✅ 可选 | ❌ 不支持 | ✅ 可选 | ✅ 可选 | ✅ 可选 | ❌ 不支持 | ❌ 不支持 | ❌ 不支持 |



#### 接口示例

##### 获取商品列表

```
# 超级管理员(root) - 可以查看所有店铺的商品
curl "http://127.0.0.1:8009/admin/product/list?page=1&page_size=10" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 超级管理员(root) - 可以查看指定店铺的商品
curl "http://127.0.0.1:8009/admin/product/list?page=1&page_size=10&shop_id=2" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 支持商品名称模糊查询
curl "http://127.0.0.1:8009/admin/product/list?page=1&page_size=10&name=油漆" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 支持店铺+名称组合查询
curl "http://127.0.0.1:8009/admin/product/list?page=1&page_size=10&shop_id=1&name=涂料" \
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
    "shop_id": 1,
    "cost": 100,
    "shipping_cost": 10,
    "product_cost": 90
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
    "remark": "",
    "cost": 100,
    "shipping_cost": 10,
    "product_cost": 90
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

**说明：**
- 支持部分字段更新，前端传什么字段就更新什么字段，不传的字段保持不变
- 支持更新的字段：`seller_price`（售价）、`specification`（规格）、`is_on_shelf`（上架状态）、`remark`（备注）、`stock`（库存）
- 不支持更新：`name`（商品名称）、`image`（商品图片）、`category_id`（分类ID）、成本相关字段
- 成本相关字段由入库操作自动更新，不支持手动修改

```bash
# 编辑商品（只更新售价和上架状态）
curl -X PUT "http://127.0.0.1:8009/admin/product/edit/4" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "seller_price": 120,
    "is_on_shelf": 1,
    "shop_id": 1
  }'

# 编辑商品（只更新库存）
curl -X PUT "http://127.0.0.1:8009/admin/product/edit/4" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stock": 100,
    "shop_id": 1
  }'

# 编辑商品（更新多个字段）
curl -X PUT "http://127.0.0.1:8009/admin/product/edit/4" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "seller_price": 120,
    "specification": "1L装",
    "is_on_shelf": 1,
    "remark": "高质量油漆",
    "stock": 50,
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
- **权限校验**：系统会验证传入的 `shop_id` 与JWT token中的权限是否匹配
  - 超级管理员(root)可以操作任意店铺的入库
  - 普通管理员(lizengchun/zhangweiyang)只能操作自己店铺的入库

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

**说明：**
- **权限校验**：系统会验证传入的 `shop_id` 与JWT token中的权限是否匹配
  - 超级管理员(root)可以操作任意店铺的出库
  - 普通管理员(lizengchun/zhangweiyang)只能操作自己店铺的出库

```bash
# 普通管理员(lizengchun) - 操作自己店铺的出库
curl -X POST "http://127.0.0.1:8009/admin/stock/batch/outbound" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
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
  }'

# 超级管理员(root) - 操作指定店铺的出库
curl -X POST "http://127.0.0.1:8009/admin/stock/batch/outbound" \
  -H "Authorization: Bearer ROOT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
        {
            "product_id": 5,
            "quantity": 1,
            "unit_price": 100,
            "total_price": 100,
            "remark": "涞水店出库"
        }
    ],
    "total_amount": 100,
    "user_name": "王五",
    "user_id": 1003,
    "operator": "root",
    "operator_id": 1,
    "shop_id": 2,
    "remark": "涞水店操作"
  }'

# 普通管理员尝试操作其他店铺出库会返回403错误
curl -X POST "http://127.0.0.1:8009/admin/stock/batch/outbound" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
        {
            "product_id": 5,
            "quantity": 1,
            "unit_price": 100,
            "total_price": 100,
            "remark": "其他店铺出库"
        }
    ],
    "total_amount": 100,
    "user_name": "王五",
    "user_id": 1003,
    "operator": "lizengchun",
    "operator_id": 2,
    "shop_id": 2,
    "remark": "其他店铺操作"
  }'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**字段说明：**
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

**说明：**
- 支持 `shop_id` 参数进行店铺筛选
- 超级管理员(root)可以查看所有店铺的库存操作，普通管理员只能查看自己店铺的库存操作
- 如果未传递 `shop_id` 参数，则自动使用JWT token中的店铺ID

```bash
# 超级管理员(root) - 查看所有库存操作
curl "http://127.0.0.1:8009/admin/stock/operations?page=1&page_size=10" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 超级管理员(root) - 查看指定店铺的库存操作
curl "http://127.0.0.1:8009/admin/stock/operations?page=1&page_size=10&shop_id=2" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 普通管理员(lizengchun) - 只能查看自己店铺的库存操作
curl "http://127.0.0.1:8009/admin/stock/operations?page=1&page_size=10" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"

# 普通管理员(lizengchun) - 尝试查看其他店铺库存操作会返回403错误
curl "http://127.0.0.1:8009/admin/stock/operations?page=1&page_size=10&shop_id=2" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**查询参数：**
- `page`: 页码，默认为1
- `page_size`: 每页大小，默认为10
- `types`: 操作类型（可选），1-入库，2-出库，3-退货
- `shop_id`: 店铺ID（可选），用于筛选特定店铺的库存操作

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

**说明：**
- 无需传递 `shop_id` 参数，系统会从JWT token中获取管理员权限
- 系统会先查询操作信息，然后验证管理员是否有权限查看该操作
- 超级管理员(root)可以查看任意店铺的库存操作详情，普通管理员只能查看自己店铺的库存操作详情

```bash
# 普通管理员(lizengchun) - 查看自己店铺的库存操作详情
curl "http://127.0.0.1:8009/admin/stock/operation/123" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"

# 超级管理员(root) - 查看任意店铺的库存操作详情
curl "http://127.0.0.1:8009/admin/stock/operation/456" \
  -H "Authorization: Bearer ROOT_TOKEN"

# 普通管理员尝试查看其他店铺库存操作详情会返回403错误
curl "http://127.0.0.1:8009/admin/stock/operation/456" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN"
# 返回: {"code":-1,"message":"无权限查看该库存操作"}
```

**请求示例：**
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

**说明：**
- 需要传递 `shop_id` 参数指定店铺
- 超级管理员(root)可以更新任意店铺的出库单支付状态，普通管理员只能更新自己店铺的出库单支付状态
- 系统会验证 `shop_id` 与管理员权限是否匹配

```bash
# 普通管理员(lizengchun) - 更新自己店铺的出库单支付状态
curl -X POST "http://127.0.0.1:8009/admin/stock/set/payment-status" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operation_id": 123,
    "payment_finish_status": 3,
    "operator": "lizengchun",
    "operator_id": 2,
    "shop_id": 1
  }'

# 超级管理员(root) - 更新指定店铺的出库单支付状态
curl -X POST "http://127.0.0.1:8009/admin/stock/set/payment-status" \
  -H "Authorization: Bearer ROOT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operation_id": 456,
    "payment_finish_status": 3,
    "operator": "root",
    "operator_id": 1,
    "shop_id": 2
  }'

# 普通管理员尝试更新其他店铺出库单支付状态会返回403错误
curl -X POST "http://127.0.0.1:8009/admin/stock/set/payment-status" \
  -H "Authorization: Bearer LIZENGCHUN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "operation_id": 456,
    "payment_finish_status": 3,
    "operator": "lizengchun",
    "operator_id": 2,
    "shop_id": 2
  }'
# 返回: {"code":-1,"message":"无权限操作该店铺的数据"}
```

**请求示例：**
```http
POST /admin/stock/set/payment-status
Content-Type: application/json

{
  "operation_id": 123,
  "payment_finish_status": 3,
  "operator": "管理员",
  "operator_id": 1001,
  "shop_id": 1
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

#### 6. 获取库存操作明细列表

**说明：**
- 获取库存操作明细列表，支持按店铺和商品筛选
- **权限校验**：系统会验证传入的 `shop_id` 与JWT token中的权限是否匹配
  - 超级管理员(root)可以查看任意店铺的库存操作明细
  - 普通管理员(lizengchun/zhangweiyang)只能查看自己店铺的库存操作明细
- 支持分页查询和按商品ID筛选

**请求参数：**
- `page`: 页码（可选，默认1）
- `page_size`: 每页数量（可选，默认10）
- `shop_id`: 店铺ID（可选，用于筛选特定店铺的明细）
- `product_id`: 商品ID（可选，用于筛选特定商品的明细）

```bash
# 获取库存操作明细列表
curl "http://127.0.0.1:8009/admin/stock/items?page=1&page_size=10&shop_id=1&product_id=2" \
  -H "Authorization: Bearer ROOT_TOKEN"
```

**请求示例：**
```http
GET /admin/stock/items?page=1&page_size=10&shop_id=1&product_id=2
```

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "operation_id": 1001,
        "shop_id": 1,
        "product_id": 2,
        "quantity": 20,
        "unit_price": 0,
        "total_price": 1320,
        "before_stock": 50,
        "after_stock": 70,
        "product_cost": 66,
        "profit": 0,
        "product_name": "华润漆",
        "specification": "5L/桶",
        "unit": "桶",
        "remark": "0901入库",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  },
  "message": "获取库存操作明细成功"
}
```

**字段说明：**
- `operation_id`: 关联的库存操作主表ID
- `shop_id`: 关联的店铺ID
- `product_id`: 商品ID
- `quantity`: 操作数量
- `unit_price`: 单价（入库时为0，出库时为售价）
- `total_price`: 总价
- `before_stock`: 操作前库存
- `after_stock`: 操作后库存
- `product_cost`: 商品成本（进价）
- `profit`: 利润（出库时计算）
- `product_name`: 商品名称
- `specification`: 商品规格
- `unit`: 商品单位
- `remark`: 备注

#### 7. 获取供货商列表

**说明：**
- 无需传递任何参数，返回所有供货商列表
- 所有管理员都可以查看完整的供货商列表

```bash
# 获取供货商列表
curl "http://127.0.0.1:8009/admin/stock/suppliers" \
  -H "Authorization: Bearer ROOT_TOKEN"
```

**请求示例：**
```http
GET /admin/stock/suppliers
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



## 需初始化的数据库表结构

### Admin

1. operator表、shop表、category表、supplier表、user表



## 测试用例Case

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



