# Lighter iOS/Android Mobile SDK

This package provides gomobile-compatible bindings for the Lighter trading platform, allowing you to integrate Lighter into iOS and Android apps.

## Building

### iOS

```bash
cd mobile
./build_ios.sh
```

This will create `../build/Lighter.xcframework` that you can use in your iOS project.

### Android

```bash
cd mobile
gomobile bind -target=android -o ../build/lighter.aar -v ./
```

## iOS Integration

1. Drag `Lighter.xcframework` into your Xcode project
2. In target settings → General → Frameworks, Libraries, and Embedded Content
3. Set the framework to "Embed & Sign"
4. Import in your Swift files: `import Lighter`

## Swift Usage Examples

### Generate API Keys

```swift
import Lighter

// Generate random API key
let result = MobileGenerateAPIKey("")
if result?.error == "" {
    print("Private Key: \(result!.privateKey!)")
    print("Public Key: \(result!.publicKey!)")
} else {
    print("Error: \(result!.error!)")
}

// Generate with seed
let seededResult = MobileGenerateAPIKey("my-secret-seed")
```

### Create Client

```swift
// Create a client for trading
let error = MobileCreateClient(
    "https://api.lighter.xyz",           // API URL
    "0x1234...",                          // Private key (hex)
    42,                                    // Chain ID
    0,                                     // API key index
    123                                    // Account index
)

if error == "" {
    print("✅ Client created successfully!")
} else {
    print("❌ Error: \(error)")
}
```

### Check Client Configuration

```swift
let checkError = MobileCheckClient(0, 123)  // apiKeyIndex, accountIndex
if checkError == "" {
    print("✅ Client verified!")
}
```

### Create an Order

```swift
let orderResult = MobileSignCreateOrder(
    0,          // marketIndex
    1,          // clientOrderIndex
    1000,       // baseAmount
    50000,      // price
    0,          // isAsk (0 = buy, 1 = sell)
    0,          // orderType (0 = limit, 1 = market, etc.)
    0,          // timeInForce
    0,          // reduceOnly
    0,          // triggerPrice
    -1,         // orderExpiry (-1 for default)
    -1          // nonce (-1 for automatic)
)

if orderResult?.error == "" {
    // Parse the JSON transaction
    if let jsonData = orderResult?.json?.data(using: .utf8) {
        let tx = try? JSONSerialization.jsonObject(with: jsonData)
        print("Transaction: \(tx)")
        // Send to Lighter API
    }
} else {
    print("Error: \(orderResult?.error ?? "unknown")")
}
```

### Cancel an Order

```swift
let cancelResult = MobileSignCancelOrder(
    0,      // marketIndex
    12345,  // orderIndex
    -1      // nonce (-1 for automatic)
)

if cancelResult?.error == "" {
    print("Cancel transaction: \(cancelResult!.json!)")
}
```

### Withdraw USDC

```swift
let withdrawResult = MobileSignWithdraw(
    1000000,  // usdcAmount (in base units)
    -1        // nonce
)

if withdrawResult?.error == "" {
    print("Withdraw transaction: \(withdrawResult!.json!)")
}
```

### Transfer Between Accounts

```swift
// Memo must be exactly 32 bytes
let memo = "Transfer memo text          "  // Pad to 32 bytes

let transferResult = MobileSignTransfer(
    456,        // toAccountIndex
    500000,     // usdcAmount
    100,        // fee
    memo,       // 32-byte memo
    -1          // nonce
)

if transferResult?.error == "" {
    let json = transferResult!.json!
    // Parse JSON and send to API
}
```

### Create Authentication Token

```swift
// Create auth token (default 7 hours expiry)
let tokenResult = MobileCreateAuthToken(0)

if tokenResult?.error == "" {
    let authToken = tokenResult!.json!
    print("Auth token: \(authToken)")
    // Use in API requests
}

// Or with custom deadline (Unix timestamp)
let customDeadline = Int64(Date().timeIntervalSince1970) + 3600  // 1 hour
let tokenResult2 = MobileCreateAuthToken(customDeadline)
```

### Update Leverage

```swift
let leverageResult = MobileSignUpdateLeverage(
    0,      // marketIndex
    5000,   // initialMarginFraction (50% = 5000, represents 0.5 with 4 decimals)
    0,      // marginMode
    -1      // nonce
)
```

### Multiple API Keys

```swift
// Create first client (API key index 0)
MobileCreateClient("https://api.lighter.xyz", privateKey1, 42, 0, 123)

// Create second client (API key index 1)
MobileCreateClient("https://api.lighter.xyz", privateKey2, 42, 1, 123)

// Switch between them
MobileSwitchAPIKey(0)  // Use first key
let order1 = MobileSignCreateOrder(/* ... */)

MobileSwitchAPIKey(1)  // Use second key
let order2 = MobileSignCreateOrder(/* ... */)
```

## Complete iOS Example

```swift
import UIKit
import Lighter

class TradingViewController: UIViewController {
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupLighterClient()
    }
    
    func setupLighterClient() {
        // Generate or load API keys
        let apiKey = MobileGenerateAPIKey("")
        guard let privateKey = apiKey?.privateKey, apiKey?.error == "" else {
            print("Failed to generate key")
            return
        }
        
        // Create client
        let error = MobileCreateClient(
            "https://api.lighter.xyz",
            privateKey,
            42,     // Chain ID
            0,      // API key index
            123     // Account index
        )
        
        guard error == "" else {
            print("Failed to create client: \(error)")
            return
        }
        
        print("✅ Lighter client ready!")
    }
    
    func placeOrder() {
        let result = MobileSignCreateOrder(
            0,          // BTC-USDC market (example)
            Int64(Date().timeIntervalSince1970),  // unique client order ID
            100000000,  // 1 BTC (assuming 8 decimals)
            4500000,    // $45,000 price
            0,          // Buy order
            0,          // Limit order
            0,          // Good till cancelled
            0,          // Not reduce-only
            0,          // No trigger
            -1,         // Default expiry
            -1          // Auto nonce
        )
        
        guard let txResult = result, txResult.error == "" else {
            print("Error creating order: \(result?.error ?? "unknown")")
            return
        }
        
        // Send transaction to Lighter API
        sendToLighterAPI(json: txResult.json!)
    }
    
    func sendToLighterAPI(json: String) {
        // Your API call implementation
        guard let url = URL(string: "https://api.lighter.xyz/transactions") else { return }
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = json.data(using: .utf8)
        
        URLSession.shared.dataTask(with: request) { data, response, error in
            if let error = error {
                print("API Error: \(error)")
                return
            }
            print("✅ Transaction submitted!")
        }.resume()
    }
}
```

## Function Reference

### Key Management
- `MobileGenerateAPIKey(seed: String) -> APIKeyResult?` - Generate API key pair
- `MobileCreateClient(url, privateKey: String, chainId, apiKeyIndex: Int, accountIndex: Int64) -> String` - Create client (returns error string, empty on success)
- `MobileCheckClient(apiKeyIndex: Int, accountIndex: Int64) -> String` - Verify client
- `MobileSwitchAPIKey(apiKeyIndex: Int) -> String` - Switch active API key

### Trading Operations
- `MobileSignCreateOrder(...) -> TxResult?` - Create order transaction
- `MobileSignCancelOrder(marketIndex: Int, orderIndex, nonce: Int64) -> TxResult?` - Cancel order
- `MobileSignCancelAllOrders(timeInForce: Int, time, nonce: Int64) -> TxResult?` - Cancel all orders
- `MobileSignModifyOrder(...) -> TxResult?` - Modify existing order

### Account Management
- `MobileSignWithdraw(usdcAmount, nonce: Int64) -> TxResult?` - Withdraw USDC
- `MobileSignTransfer(...) -> TxResult?` - Transfer between accounts
- `MobileSignCreateSubAccount(nonce: Int64) -> TxResult?` - Create sub-account
- `MobileSignChangePubKey(pubKey: String, nonce: Int64) -> TxResult?` - Change public key

### Pool Operations
- `MobileSignCreatePublicPool(...) -> TxResult?` - Create liquidity pool
- `MobileSignUpdatePublicPool(...) -> TxResult?` - Update pool settings
- `MobileSignMintShares(...) -> TxResult?` - Mint pool shares
- `MobileSignBurnShares(...) -> TxResult?` - Burn pool shares

### Position Management
- `MobileSignUpdateLeverage(...) -> TxResult?` - Update leverage
- `MobileSignUpdateMargin(...) -> TxResult?` - Update margin

### Authentication
- `MobileCreateAuthToken(deadline: Int64) -> TxResult?` - Create auth token

## Return Types

### APIKeyResult
```swift
class APIKeyResult {
    var privateKey: String?
    var publicKey: String?
    var error: String?
}
```

### TxResult
```swift
class TxResult {
    var json: String?    // Transaction JSON to send to API
    var error: String?   // Error message if failed
}
```

## Error Handling

All transaction functions return a `TxResult` object. Always check the `error` field:

```swift
let result = MobileSignCreateOrder(/* ... */)
if result?.error != "" {
    // Handle error
    print("Error: \(result?.error ?? "unknown")")
} else {
    // Success - use result.json
    processTransaction(result!.json!)
}
```

## Notes

- All nonce parameters: use `-1` for automatic nonce management
- Order expiry: use `-1` for default (28 days)
- All amounts are in base units (check Lighter API docs for decimals)
- Memo fields must be exactly 32 bytes
- Private keys must be hex-encoded with `0x` prefix

## Android Usage

The same functions are available in Android with Kotlin/Java:

```kotlin
import mobile.Mobile

// Generate key
val result = Mobile.generateAPIKey("")
println("Private Key: ${result.privateKey}")

// Create client
val error = Mobile.createClient(
    "https://api.lighter.xyz",
    result.privateKey,
    42,
    0,
    123
)
```

## Support

For issues or questions, visit: https://github.com/elliottech/lighter-go

