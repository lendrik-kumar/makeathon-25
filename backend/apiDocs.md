# Second-Hand Product Verification System API Documentation

This API enables users to create, manage, and verify the authenticity and history of products throughout their lifecycle.

## User Management

### 1. Register User
**POST /api/users/register**

Creates a new user account with role-based permissions.

**Request Body:**
```json
{
  "username": "john_smith",
  "password": "secure_password",
  "role": "regular"
}
```
*Roles:* `regular` (buyer/seller), `brand` (manufacturer), `repair_shop`, `admin`

**Response:**
```json
{
  "message": "User registered"
}
```

**User Value:** Account creation is the gateway to using the system. Different roles provide different capabilities - manufacturers can register products, repair shops can log repairs, and regular users can transfer ownership.

### 2. User Login
**POST /api/users/login**

Authenticates a user and provides an access token.

**Request Body:**
```json
{
  "username": "john_smith",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**User Value:** Provides secure access to protected features based on the user's role.

## Product Management

### 3. Register Product
**POST /api/products**

Registers a new product in the system, creating its first blockchain-like record.

**Request Body:**
```json
{
  "serial_number": "SN12345678",
  "manufacturer": "TechCorp",
  "model": "UltraPhone X"
}
```

**Response:**
```json
{
  "id": 1,
  "serial_number": "SN12345678",
  "manufacturer": "TechCorp",
  "model": "UltraPhone X",
  "created_at": "2025-04-26T10:30:00Z",
  "updated_at": "2025-04-26T10:30:00Z"
}
```

**User Value:** Manufacturers can add new products to the system, establishing the start of its verifiable history.

### 4. Get Product Details
**GET /api/products/:id**

Retrieves complete information about a product including its ownership history.

**Response:**
```json
{
  "product": {
    "id": 1,
    "serial_number": "SN12345678",
    "manufacturer": "TechCorp",
    "model": "UltraPhone X"
  },
  "history": [
    {
      "id": 1,
      "event_type": "registration",
      "event_data": "{\"details\": \"Product registered\"}",
      "created_at": "2025-04-26T10:30:00Z",
      "created_by": 5
    },
    {
      "id": 2,
      "event_type": "ownership_transfer",
      "event_data": "{\"new_owner_id\": 8}",
      "created_at": "2025-04-27T14:15:00Z",
      "created_by": 5
    }
  ]
}
```

**User Value:** Buyers can verify a product's complete history before purchasing, seeing all repairs, transfers, and other events.

### 5. Create Product Event
**POST /api/products/:id/events**

Records important events in a product's lifecycle (repair, warranty claim, etc.).

**Request Body:**
```json
{
  "event_type": "repair",
  "event_data": "{\"repair_details\": \"Screen replacement\", \"parts_used\": \"Original screen\"}"
}
```

**Response:**
```json
{
  "id": 3,
  "product_id": 1,
  "event_type": "repair",
  "event_data": "{\"repair_details\": \"Screen replacement\", \"parts_used\": \"Original screen\"}",
  "created_at": "2025-04-28T09:45:00Z",
  "created_by": 12,
  "event_hash": "8f7d9a6c5b4e3d2f1a0b9c8d7e6f5a4b3c2d1e0f",
  "previous_event_hash": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
}
```

**User Value:** Repair shops can document their work, manufacturers can update warranty information, and owners can record maintenance.

### 6. Initiate Ownership Transfer
**POST /api/products/:id/transfer**

Starts the process of transferring ownership to another user.

**Request Body:**
```json
{
  "new_owner_username": "alice_jones"
}
```

**Response:**
```json
{
  "message": "Transfer initiated"
}
```

**User Value:** Sellers can safely begin the process of transferring product ownership to a buyer.

### 7. Confirm Ownership Transfer
**POST /api/products/:id/transfer/confirm**

The recipient confirms acceptance of ownership transfer.

**Response:**
```json
{
  "message": "Transfer confirmed"
}
```

**User Value:** Buyers can approve the transfer, establishing themselves as the new verified owner in the product's history.

### 8. Verify Product History
**GET /api/products/:id/verify**

Validates the integrity of a product's entire event chain.

**Response:**
```json
{
  "message": "History is valid"
}
```

**User Value:** Provides tamper-proof verification that a product's history hasn't been altered, giving buyers confidence in the authenticity of second-hand items.

## User Journey Examples

### For Manufacturers
1. Register as a brand
2. Register new products
3. Record warranty information

### For Repair Shops
1. Register as a repair shop
2. Record repair events on products
3. Provide verifiable repair history

### For Buyers/Sellers
1. Register as a regular user
2. View product history before purchasing
3. Initiate/confirm ownership transfers
4. Verify product authenticity