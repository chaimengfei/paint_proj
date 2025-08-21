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

## API接口说明

### 商品管理接口

##### 新增商品
```http
POST /admin/product/add
Content-Type: application/json

{
  "name": "调和漆",
  "category_id": 1,
  "image": "https://example.com/paint.jpg",
  "seller_price": 5000,
  "cost": 3000,
  "shipping_cost": 500,
  "product_cost": 2500,
  "specification": "3kg*4",
  "unit": "L",
  "remark": "优质调和漆",
  "is_on_shelf": 1
}
```

##### 编辑商品
```http
PUT /admin/product/edit/1
Content-Type: application/json

{
  "name": "调和漆",
  "category_id": 1,
  "image": "https://example.com/paint.jpg",
  "seller_price": 5500,
  "cost": 3000,
  "shipping_cost": 500,
  "product_cost": 2500,
  "specification": "3kg*4",
  "unit": "L",
  "remark": "优质调和漆",
  "is_on_shelf": 1
}
```

##### 删除商品
```http
DELETE /admin/product/delete/1
```

##### 获取商品列表
```http
GET /admin/product/list?page=1&page_size=10
```

##### 根据ID获取商品信息
```http
GET /admin/product/1
```

响应示例：
```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "调和漆",
    "category_id": 1,
    "image": "https://example.com/paint.jpg",
    "seller_price": 5000,
    "cost": 3000,
    "shipping_cost": 500,
    "product_cost": 2500,
    "specification": "3kg*4",
    "unit": "L",
    "remark": "优质调和漆",
    "stock": 100,
    "is_on_shelf": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
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
      "unit_price": 3000,
      "remark": "优质调和漆",
      "product_name": "",
      "specification": "",
      "unit": "",
      "total_price": 0
    },
    {
      "product_id": 2,
      "quantity": 20,
      "unit_price": 2500,
      "remark": "标准规格",
      "product_name": "",
      "specification": "",
      "unit": "",
      "total_price": 0
    }
  ],
  "total_amount": 80000,
  "operator": "张三",
  "operator_id": 1001,
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
GET /admin/stock/operations?page=1&page_size=10
```

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

#### 字段说明

**批量入库请求字段：**
- `items`: 入库商品列表
  - `product_id`: 商品ID（必填）
  - `quantity`: 入库数量（必填）
  - `unit_price`: 单价（可选，单位：分）
  - `remark`: 备注（可选）
  - `product_name`: 商品全名（自动补齐，前端可传空字符串）
  - `specification`: 规格（自动补齐，前端可传空字符串）
  - `unit`: 单位（自动补齐，前端可传空字符串）
  - `total_price`: 总金额（自动计算，前端可传0）
- `total_amount`: 总金额（前端计算，单位：分）
- `operator`: 操作人姓名（必填）
- `operator_id`: 操作人ID（必填）
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
- `type`: 1-入库, 2-出库, 3-退货
- `operator_type`: 1-用户, 2-系统, 3-管理员

**说明：**
- 后端会自动补齐商品信息（`product_name`, `specification`, `unit`, `total_price`）
- 前端在items中传入这些字段时可以使用空字符串或0，后端会自动填充
- 如果没有提供 `unit_price`，会使用商品的 `seller_price`
- 时间字段由后端自动记录，无需前端传入

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
  remark VARCHAR(500) DEFAULT '' COMMENT '备注',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);
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

## 使用说明

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