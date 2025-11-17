# Lighter SDK - Quick Reference for Frontend

## 🚀 Quick Start (5 minutes)

### 1. Setup (One time)
```bash
# Copy to your Expo project
cp -r expo-lighter-module /path/to/your-app/

# Install
npm install
npx expo prebuild
npx expo run:ios
```

### 2. Basic Usage
```typescript
import { LighterSDK } from 'expo-lighter-module';

// Generate keys
const keys = LighterSDK.generateAPIKey('');

// Create client
LighterSDK.createClient('https://api.lighter.xyz', keys.privateKey, 42, 0, 123);

// Place order
const order = LighterSDK.signCreateOrder({
  marketIndex: 0,
  clientOrderIndex: Date.now(),
  baseAmount: 100000000,
  price: 5000000,
  isAsk: 0, // 0=buy, 1=sell
  orderType: 0, // 0=limit
  timeInForce: 0,
  reduceOnly: 0,
  triggerPrice: 0
});

// Send to API
fetch('https://api.lighter.xyz/transactions', {
  method: 'POST',
  body: order.json
});
```

## 📚 All Functions

### Key Management
| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `generateAPIKey(seed)` | `seed: string` | `{privateKey, publicKey, error}` | Generate new key pair |
| `createClient(url, key, chainId, apiIdx, accIdx)` | 5 params | `error: string` | Initialize client |
| `checkClient(apiIdx, accIdx)` | 2 params | `error: string` | Verify client |
| `switchAPIKey(apiIdx)` | `apiIdx: number` | `error: string` | Switch keys |

### Trading
| Function | Returns | Description |
|----------|---------|-------------|
| `signCreateOrder({...})` | `{json, error}` | Create new order |
| `signCancelOrder(market, orderId, nonce)` | `{json, error}` | Cancel order |
| `signCancelAllOrders(tif, time, nonce)` | `{json, error}` | Cancel all |
| `signModifyOrder({...})` | `{json, error}` | Modify order |

### Account
| Function | Returns | Description |
|----------|---------|-------------|
| `signWithdraw(amount, nonce)` | `{json, error}` | Withdraw USDC |
| `signTransfer({...})` | `{json, error}` | Transfer USDC |
| `signCreateSubAccount(nonce)` | `{json, error}` | New sub-account |
| `signChangePubKey(pubKey, nonce)` | `{json, error}` | Change key |
| `createAuthToken(deadline)` | `{json, error}` | Get JWT token |

## 🎯 Common Patterns

### Error Handling
```typescript
const result = LighterSDK.signCreateOrder({...});
if (result.error !== '') {
  // Handle error
  console.error(result.error);
  return;
}
// Use result.json
```

### Creating Orders
```typescript
// Buy 1 BTC at $50,000
LighterSDK.signCreateOrder({
  marketIndex: 0,              // BTC market
  clientOrderIndex: Date.now(), // Unique ID
  baseAmount: 100000000,       // 1.0 (8 decimals)
  price: 5000000,              // $50,000 
  isAsk: 0,                    // Buy
  orderType: 0,                // Limit
  timeInForce: 0,              // GTC
  reduceOnly: 0,               // No
  triggerPrice: 0,             // No trigger
  orderExpiry: -1,             // Default (28 days)
  nonce: -1                    // Auto
});
```

### Withdraw Money
```typescript
// Withdraw $100 USDC
const result = LighterSDK.signWithdraw(
  100000000,  // $100 (6 decimals)
  -1          // Auto nonce
);
```

### Transfer to Another User
```typescript
const memo = 'Payment'.padEnd(32, ' '); // MUST be 32 bytes!

const result = LighterSDK.signTransfer({
  toAccountIndex: 456,
  usdcAmount: 50000000, // $50
  fee: 100,
  memo: memo,
  nonce: -1
});
```

### Authentication
```typescript
// Get auth token for API calls
const result = LighterSDK.createAuthToken(0); // 0 = 7 hour expiry
const token = result.json;

// Use in requests
fetch(url, {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

## 🔢 Constants

```typescript
// Order Side
const BUY = 0;
const SELL = 1;

// Order Type
const LIMIT = 0;
const MARKET = 1;
const POST_ONLY = 2;
const FILL_OR_KILL = 3;
const IMMEDIATE_OR_CANCEL = 4;

// Time In Force
const GTC = 0; // Good Till Cancelled
const IOC = 1; // Immediate Or Cancel
const FOK = 2; // Fill Or Kill
const PO = 3;  // Post Only

// Margin Direction
const ADD_MARGIN = 0;
const REMOVE_MARGIN = 1;
```

## 💡 Tips

1. **Always use -1 for nonce** (automatic)
2. **Always use -1 for orderExpiry** (default 28 days)
3. **Memo must be 32 bytes** - use `.padEnd(32, ' ')`
4. **Check error field** before using result
5. **Create client once** per session
6. **Cache auth tokens** (valid for hours)
7. **Use Date.now()** for unique order IDs

## ⚠️ Common Errors

| Error | Solution |
|-------|----------|
| "client is not created" | Call `createClient()` first |
| "invalid account index" | accountIndex must be > 0 |
| "memo expected to be 32 bytes" | Use `.padEnd(32, ' ')` |
| "invalid pub key length" | Public key must be 40 bytes |

## 📱 React Native Example

```typescript
import { useState, useEffect } from 'react';
import { LighterSDK } from 'expo-lighter-module';

function TradingScreen() {
  const [keys, setKeys] = useState(null);
  
  useEffect(() => {
    // Initialize on mount
    const apiKeys = LighterSDK.generateAPIKey('');
    if (apiKeys.error === '') {
      setKeys(apiKeys);
      
      const err = LighterSDK.createClient(
        'https://api.lighter.xyz',
        apiKeys.privateKey,
        42,
        0,
        123
      );
      
      if (err === '') {
        console.log('✅ Ready to trade');
      }
    }
  }, []);
  
  const buyBTC = async () => {
    const order = LighterSDK.signCreateOrder({
      marketIndex: 0,
      clientOrderIndex: Date.now(),
      baseAmount: 100000000, // 1 BTC
      price: 5000000,        // $50k
      isAsk: 0,              // Buy
      orderType: 0,
      timeInForce: 0,
      reduceOnly: 0,
      triggerPrice: 0
    });
    
    if (order.error !== '') {
      alert(order.error);
      return;
    }
    
    const response = await fetch('https://api.lighter.xyz/transactions', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: order.json
    });
    
    if (response.ok) {
      alert('Order placed!');
    }
  };
  
  return <Button title="Buy BTC" onPress={buyBTC} />;
}
```

## 🔐 Security Checklist

- [ ] Store private keys in `expo-secure-store`
- [ ] Never log private keys
- [ ] Use environment variables for URLs
- [ ] Validate user inputs
- [ ] Require confirmation for withdrawals/transfers
- [ ] Show transaction details before signing
- [ ] Use testnet first

## 📞 Support

- Full API Spec: `FRONTEND_API_SPEC.json`
- Setup Guide: `SETUP.md`
- Examples: `README.md`
- Issues: https://github.com/elliottech/lighter-go/issues

