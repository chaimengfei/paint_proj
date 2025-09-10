# 油漆销售管理系统

## 功能特性

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

## API接口说明

### 商品管理接口

#### 注意事项

1. **金额处理**: 所有金额字段在JSON中显示为元，但系统内部存储为分
2. **必填字段**: 新增和编辑商品时，`name`、`category_id`、`image`为必填字段
3. **图片上传**: 图片上传接口返回的是完整的URL地址
4. **分页参数**: 页码从1开始，每页大小默认为10
5. **商品状态**: `is_on_shelf`字段控制商品是否上架，1表示上架，0表示下架

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
curl http://127.0.0.1:8009/admin/product/list
{"code":0,"data":{"list":[{"id":1,"name":"贸彩1K白","seller_price":112.00,"cost":87.00,"shipping_cost":2.00,"product_cost":85.00,"category_id":1,"stock":100,"image":"http://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/07GE2k1DJWhah4QA_RlY91685434479136.jpg","specification":"4L","unit":"桶","remark":"","is_on_shelf":1},{"id":2,"name":"贸彩1K特黑","seller_price":108.00,"cost":67.00,"shipping_cost":0.00,"product_cost":67.00,"category_id":1,"stock":118,"image":"https://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/eBJ1AAdnMvKC-MiM1MK0O1686032850008.png","specification":"8L","unit":"桶","remark":"","is_on_shelf":1},{"id":3,"name":"原子灰","seller_price":85.00,"cost":80.00,"shipping_cost":0.00,"product_cost":80.00,"category_id":3,"stock":98,"image":"https://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/eBJ1AAdnMvKC-MiM1MK0O1686032850008.png","specification":"3KG*4","unit":"箱","remark":"","is_on_shelf":1},{"id":4,"name":"贸彩2K特白","seller_price":115.00,"cost":87.00,"shipping_cost":3.00,"product_cost":85.00,"category_id":1,"stock":120,"image":"https://dsers-affiliate.s3.amazonaws.com/logo-big.png","specification":"20L","unit":"桶","remark":"","is_on_shelf":0}],"page":1,"page_size":10,"total":4}}%  
```



##### 新增商品

```http
curl --location 'http://127.0.0.1:8009/admin/product/add' \
--header 'Content-Type: application/json' \
--data '{
                "name": "贸彩1K白",
                "category_id": 1,
                "seller_price": 120,
                "image": "http://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/07GE2k1DJWhah4QA_RlY91685434479136.jpg",
                "unit": "桶",
                "is_on_shelf":1,
                "remark": ""
            }'
{"code":0,"message":"添加成功"}% 
```

##### 编辑商品
```http
 curl --location --request PUT 'http://127.0.0.1:8009/admin/product/edit/4' \
--header 'Content-Type: application/json' \
--data '{
                "id":1,
                "name": "贸彩1K白",
                "category_id": 1,
                "seller_price": 120,
                "image": "http://dsers-dev-public.oss-cn-zhangjiakou.aliyuncs.com/07GE2k1DJWhah4QA_RlY91685434479136.jpg",
                "unit": "桶",
                "is_on_shelf":1,
                "remark": ""
            }'
{"code":0,"message":"编辑成功"}
```

##### 删除商品
```http
curl --location --request DELETE 'http://127.0.0.1:8009/admin/product/del/1'
```

##### 获取商品分类

```
➜  ~ curl 'http://192.168.1.6:8009/admin/product/categories'
{"code":0,"data":[{"id":1,"name":"1K-漆","sort_order":100},{"id":2,"name":"2K-漆","sort_order":99},{"id":3,"name":"辅料","sort_order":98}]}% 
```

##### 新增商品分类

```
➜  ~ curl --location 'http://192.168.1.6:8009/admin/product/category/add' \
--header 'Content-Type: application/json' \
--data '{
    "name":"测试分类1",
    "sort_order":96
}'
{"code":0,"message":"添加分类成功"}
```

##### 编辑商品分类

```
curl --location --request PUT 'http://192.168.1.6:8009/admin/product/category/edit/4' \
--header 'Content-Type: application/json' \
--data '{
    "id":4,
    "name":"测试分类4",
    "sort_order":96
}'
{ "code": 0,"message": "编辑分类成功"}
```

##### 删除商品分类

```
curl --location --request DELETE 'http://192.168.1.6:8009/admin/product/category/del/4' 
{ "code": 0,"message": "删除分类成功"}
```



### 库存管理接口

#### 1. 批量入库操作
```http
POST /admin/stock/batch/inbound
Content-Type: application/json

{
  "items": [
    {
      "product_id": 1,
      "quantity": 10,
      "cost": 3000,
      "shipping_cost": 500,
      "product_cost": 2500,
      "remark": "优质调和漆",
      "product_name": "",
      "specification": "",
      "unit": "",
      "total_amount": 0
    },
    {
      "product_id": 2,
      "quantity": 20,
      "cost": 2500,
      "shipping_cost": 400,
      "product_cost": 2100,
      "remark": "标准规格",
      "product_name": "",
      "specification": "",
      "unit": "",
      "total_amount": 0
    }
  ],
  "total_amount": 80000,
  "operator": "张三",
  "operator_id": 1001,
  "supplier": "华润涂料有限公司",
  "remark": "新货入库"
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "批量入库成功"
}
```

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
            "product_name": "固态灰",
            "unit": "箱",
            "specification": "3KG*4",
            "remark": "调鼻尖"
            
        }
    ],
    "total_amount": 170,
    "user_name": "李四",
    "user_id": 1002,
    "user_account": "13671210659",
    "operator": "张三",
    "operator_id": 1001,
    "remark": "后台操作"
}
```

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
  - `unit_price`: 单价（可选，单位：分）
  - `remark`: 备注（可选）
  - `product_name`: 商品全名（自动补齐，前端可传空字符串）
  - `specification`: 规格（自动补齐，前端可传空字符串）
  - `unit`: 单位（自动补齐，前端可传空字符串）
  - `total_price`: 总金额（自动计算，前端可传0）
- `total_amount`: 总金额（前端计算，单位：分）
- `user_name`: 用户名称（必填）
- `user_id`: 用户ID（必填）
- `user_account`: 用户账号（必填）
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
- 后端会自动补齐商品信息（`product_name`, `specification`, `unit`, `total_price`）
- 前端在items中传入这些字段时可以使用空字符串或0，后端会自动填充
- 入库时：如果没有提供 `unit_price`，会使用商品的 `seller_price`；如果新成本价更低，会自动更新商品成本价并记录变更
- 出库时：`unit_price` 为售价
- 时间字段由后端自动记录，无需前端传入



### 地址管理接口文档

---

#### 1. 获取地址列表（支持搜索）

```http
curl --location 'http://127.0.0.1:8009/admin/address/list?=null'
{"code":0,"data":[{"address_id":1,"user_id":123,"user_name":"","recipient_name":"柴五","recipient_phone":"13161621682","province":"北京市","city":"北京市","district":"东城区","detail":"123","is_default":false,"created_at":""}]}%  
```



---

#### 2. 新增地址

| 参数名  | 类型  | 必填 | 说明                           |
| ------- | ----- | ---- | ------------------------------ |
| user_id | int64 | 是   | 用户ID，指定为哪个用户创建地址 |

```http
POST /admin/address/add?user_id=123
Content-Type: application/json

{
  "data": {
    "recipient_name": "李四",
    "recipient_phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路1号",
    "is_default": true
  }
}
```

---

#### 3. 编辑地址

**路径参数**

| 参数名 | 类型  | 必填 | 说明   |
| ------ | ----- | ---- | ------ |
| id     | int64 | 是   | 地址ID |

```http
PUT /admin/address/edit/1
Content-Type: application/json

{
  "data": {
    "recipient_name": "李四",
    "recipient_phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园路2号",
    "is_default": false
  }
}
```

---

#### 4. 删除地址

**路径参数**

| 参数名 | 类型  | 必填 | 说明   |
| ------ | ----- | ---- | ------ |
| id     | int64 | 是   | 地址ID |

**查询参数**

| 参数名  | 类型  | 必填 | 说明                 |
| ------- | ----- | ---- | -------------------- |
| user_id | int64 | 是   | 用户ID，用于验证权限 |

```http
DELETE /admin/address/del/1?user_id=123
```

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

