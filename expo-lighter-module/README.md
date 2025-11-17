# Expo Lighter Module

Native iOS module for integrating Lighter trading platform into React Native apps using Expo.

## 📋 Requirements

- Expo SDK 49+
- iOS 13.0+
- Custom development build (not compatible with Expo Go)

## 🚀 Installation

### 1. Install the module

Copy the `expo-lighter-module` directory into your Expo project:

```bash
cp -r expo-lighter-module /path/to/your-expo-app/
```

### 2. Add to package.json

```json
{
  "dependencies": {
    "expo-lighter-module": "file:./expo-lighter-module"
  }
}
```

### 3. Install dependencies

```bash
npm install
# or
yarn install
```

### 4. Configure app.json

Add the module to your `app.json`:

```json
{
  "expo": {
    "plugins": [
      [
        "expo-lighter-module"
      ]
    ]
  }
}
```

### 5. Build custom development client

```bash
# iOS
npx expo prebuild -p ios
npx expo run:ios

# Or using EAS
eas build --profile development --platform ios
```

## 📖 Usage

### Import

```typescript
import { LighterSDK } from 'expo-lighter-module';
```

### Generate API Keys

```typescript
// Generate random key
const { privateKey, publicKey, error } = LighterSDK.generateAPIKey('');

if (error === '') {
  console.log('Private Key:', privateKey);
  console.log('Public Key:', publicKey);
} else {
  console.error('Error:', error);
}

// Generate with seed (deterministic)
const deterministicKey = LighterSDK.generateAPIKey('my-secret-seed');
```

### Create Client

```typescript
const error = LighterSDK.createClient(
  'https://api.lighter.xyz',     // API URL
  privateKey,                     // Your private key
  42,                             // Chain ID
  0,                              // API key index
  123                             // Account index
);

if (error === '') {
  console.log('✅ Client created successfully');
} else {
  console.error('❌ Error:', error);
}
```

### Place an Order

```typescript
const orderResult = LighterSDK.signCreateOrder({
  marketIndex: 0,
  clientOrderIndex: Date.now(),  // Unique order ID
  baseAmount: 1000000,            // Amount in base units
  price: 50000,                   // Price
  isAsk: 0,                       // 0 = buy, 1 = sell
  orderType: 0,                   // 0 = limit, 1 = market
  timeInForce: 0,                 // 0 = GTC (good till cancelled)
  reduceOnly: 0,                  // 0 = false, 1 = true
  triggerPrice: 0,                // 0 for no trigger
  orderExpiry: -1,                // -1 for default (28 days)
  nonce: -1                       // -1 for automatic
});

if (orderResult.error === '') {
  const transaction = JSON.parse(orderResult.json);
  console.log('Transaction:', transaction);
  
  // Send to Lighter API
  await sendToLighterAPI(transaction);
} else {
  console.error('Error:', orderResult.error);
}
```

### Cancel an Order

```typescript
const cancelResult = LighterSDK.signCancelOrder(
  0,       // marketIndex
  12345,   // orderIndex
  -1       // nonce (-1 for automatic)
);

if (cancelResult.error === '') {
  await sendToLighterAPI(JSON.parse(cancelResult.json));
}
```

### Withdraw USDC

```typescript
const withdrawResult = LighterSDK.signWithdraw(
  1000000,  // usdcAmount in base units
  -1        // nonce
);

if (withdrawResult.error === '') {
  await sendToLighterAPI(JSON.parse(withdrawResult.json));
}
```

### Transfer Between Accounts

```typescript
// Memo must be exactly 32 bytes
const memo = 'Transfer memo'.padEnd(32, ' ');

const transferResult = LighterSDK.signTransfer({
  toAccountIndex: 456,
  usdcAmount: 500000,
  fee: 100,
  memo: memo,
  nonce: -1
});
```

### Create Authentication Token

```typescript
// Default expiry (7 hours)
const tokenResult = LighterSDK.createAuthToken(0);

if (tokenResult.error === '') {
  const authToken = tokenResult.json;
  console.log('Auth Token:', authToken);
  
  // Use in API requests
  // Authorization: Bearer ${authToken}
}

// Custom expiry
const customDeadline = Math.floor(Date.now() / 1000) + 3600; // 1 hour
const customTokenResult = LighterSDK.createAuthToken(customDeadline);
```

## 📱 Complete React Native Example

```typescript
import React, { useEffect, useState } from 'react';
import { View, Text, Button, ActivityIndicator } from 'react-native';
import { LighterSDK } from 'expo-lighter-module';

export default function TradingScreen() {
  const [isLoading, setIsLoading] = useState(true);
  const [privateKey, setPrivateKey] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    initializeLighter();
  }, []);

  const initializeLighter = () => {
    try {
      // Generate or load API key
      const keyResult = LighterSDK.generateAPIKey('');
      
      if (keyResult.error !== '') {
        setError(keyResult.error);
        setIsLoading(false);
        return;
      }

      setPrivateKey(keyResult.privateKey);

      // Create client
      const clientError = LighterSDK.createClient(
        'https://api.lighter.xyz',
        keyResult.privateKey,
        42,
        0,
        123
      );

      if (clientError !== '') {
        setError(clientError);
      } else {
        console.log('✅ Lighter SDK initialized');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const placeOrder = async () => {
    try {
      const result = LighterSDK.signCreateOrder({
        marketIndex: 0,
        clientOrderIndex: Date.now(),
        baseAmount: 1000000,
        price: 50000,
        isAsk: 0,
        orderType: 0,
        timeInForce: 0,
        reduceOnly: 0,
        triggerPrice: 0,
        orderExpiry: -1,
        nonce: -1
      });

      if (result.error !== '') {
        setError(result.error);
        return;
      }

      const tx = JSON.parse(result.json);
      
      // Send to Lighter API
      const response = await fetch('https://api.lighter.xyz/transactions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: result.json
      });

      if (response.ok) {
        console.log('✅ Order placed successfully');
      }
    } catch (err) {
      setError(err.message);
    }
  };

  if (isLoading) {
    return <ActivityIndicator size="large" />;
  }

  return (
    <View style={{ padding: 20 }}>
      <Text style={{ fontSize: 24, marginBottom: 20 }}>Lighter Trading</Text>
      
      {error ? (
        <Text style={{ color: 'red' }}>Error: {error}</Text>
      ) : (
        <>
          <Text>Private Key: {privateKey.slice(0, 20)}...</Text>
          <Button title="Place Order" onPress={placeOrder} />
        </>
      )}
    </View>
  );
}
```

## 🔧 API Reference

### Key Management

#### `generateAPIKey(seed: string): APIKeyResult`
Generate a new API key pair.

**Parameters:**
- `seed`: Seed string for deterministic generation. Use empty string for random.

**Returns:**
```typescript
{
  privateKey: string;
  publicKey: string;
  error: string;
}
```

#### `createClient(url, privateKey, chainId, apiKeyIndex, accountIndex): string`
Create a new trading client.

**Returns:** Empty string on success, error message on failure.

#### `checkClient(apiKeyIndex, accountIndex): string`
Verify client configuration matches server.

#### `switchAPIKey(apiKeyIndex): string`
Switch between multiple API keys.

### Trading Operations

#### `signCreateOrder(params): TxResult`
Sign a create order transaction.

**Parameters:**
```typescript
{
  marketIndex: number;
  clientOrderIndex: number;
  baseAmount: number;
  price: number;
  isAsk: number;          // 0 = buy, 1 = sell
  orderType: number;       // 0 = limit, 1 = market
  timeInForce: number;
  reduceOnly: number;
  triggerPrice: number;
  orderExpiry?: number;    // -1 for default
  nonce?: number;          // -1 for automatic
}
```

#### `signCancelOrder(marketIndex, orderIndex, nonce?): TxResult`
Sign a cancel order transaction.

#### `signCancelAllOrders(timeInForce, time, nonce?): TxResult`
Sign a cancel all orders transaction.

#### `signModifyOrder(params): TxResult`
Sign a modify order transaction.

### Account Management

#### `signWithdraw(usdcAmount, nonce?): TxResult`
Sign a withdraw transaction.

#### `signTransfer(params): TxResult`
Sign a transfer transaction.

**Parameters:**
```typescript
{
  toAccountIndex: number;
  usdcAmount: number;
  fee: number;
  memo: string;    // Must be exactly 32 bytes
  nonce?: number;
}
```

#### `signCreateSubAccount(nonce?): TxResult`
Sign a create sub-account transaction.

#### `signChangePubKey(pubKey, nonce?): TxResult`
Sign a change public key transaction.

### Pool Operations

#### `signCreatePublicPool(params): TxResult`
Sign a create public pool transaction.

#### `signUpdatePublicPool(params): TxResult`
Sign an update public pool transaction.

#### `signMintShares(publicPoolIndex, shareAmount, nonce?): TxResult`
Sign a mint shares transaction.

#### `signBurnShares(publicPoolIndex, shareAmount, nonce?): TxResult`
Sign a burn shares transaction.

### Position Management

#### `signUpdateLeverage(params): TxResult`
Sign an update leverage transaction.

#### `signUpdateMargin(params): TxResult`
Sign an update margin transaction.

### Authentication

#### `createAuthToken(deadline?): TxResult`
Create an authentication token.

**Parameters:**
- `deadline`: Unix timestamp. Use 0 for default (7 hours).

## 🔒 Security Notes

- **Never commit private keys** to version control
- Store keys securely using `expo-secure-store`
- Use environment variables for API endpoints
- Validate all user inputs before signing transactions

## 🐛 Troubleshooting

### Module not found
```bash
# Clean and rebuild
rm -rf ios/build
npx expo prebuild -p ios --clean
```

### Framework not linked
Check that `Lighter.xcframework` exists in `expo-lighter-module/ios/`

### Build errors
```bash
# Reinstall pods
cd ios
pod deintegrate
pod install
```

## 📝 License

MIT

## 🤝 Support

For issues or questions:
- GitHub: https://github.com/elliottech/lighter-go
- Documentation: https://docs.lighter.xyz

