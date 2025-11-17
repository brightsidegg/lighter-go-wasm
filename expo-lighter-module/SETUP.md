# Expo Lighter Module Setup Guide

Complete step-by-step guide for integrating the Lighter SDK into your Expo React Native app.

## Prerequisites

- Node.js 18+ installed
- Expo CLI installed: `npm install -g expo-cli`
- Xcode 14+ (for iOS development)
- An existing Expo app or create a new one

## Step 1: Create or Use Existing Expo App

### New Expo App
```bash
npx create-expo-app my-lighter-app
cd my-lighter-app
```

### Existing App
```bash
cd your-existing-expo-app
```

## Step 2: Install the Module

### Option A: Copy Module Into Your Project
```bash
# From the lighter-go repository root
cp -r expo-lighter-module /path/to/your-expo-app/
```

### Option B: Use as a Local Package
```bash
# In your package.json
{
  "dependencies": {
    "expo-lighter-module": "file:../lighter-go/expo-lighter-module"
  }
}
```

## Step 3: Install Dependencies

```bash
npm install
# or
yarn install
```

## Step 4: Configure Your App

### Update app.json

Add the module plugin:

```json
{
  "expo": {
    "name": "My Lighter App",
    "slug": "my-lighter-app",
    "version": "1.0.0",
    "ios": {
      "bundleIdentifier": "com.yourcompany.lighterapp",
      "supportsTablet": true
    },
    "plugins": [
      "expo-lighter-module"
    ]
  }
}
```

## Step 5: Generate Native Projects

```bash
npx expo prebuild
```

This will create the `ios/` and `android/` directories with native code.

## Step 6: Install iOS Pods

```bash
cd ios
pod install
cd ..
```

## Step 7: Build and Run

### Development Build

```bash
# iOS Simulator
npx expo run:ios

# iOS Device (requires Apple Developer account)
npx expo run:ios --device
```

### Using EAS Build

```bash
# Install EAS CLI
npm install -g eas-cli

# Login to Expo
eas login

# Configure EAS
eas build:configure

# Create development build
eas build --profile development --platform ios

# Download and install on device
```

## Step 8: Test the Integration

Create a simple test screen:

```typescript
// App.tsx
import React, { useEffect, useState } from 'react';
import { StyleSheet, Text, View, Button, ScrollView } from 'react-native';
import { LighterSDK } from 'expo-lighter-module';

export default function App() {
  const [result, setResult] = useState('');
  const [status, setStatus] = useState('Not initialized');

  const testGeneration = () => {
    try {
      const keys = LighterSDK.generateAPIKey('');
      if (keys.error === '') {
        setResult(`Generated Keys:\nPrivate: ${keys.privateKey.slice(0, 20)}...\nPublic: ${keys.publicKey.slice(0, 20)}...`);
        setStatus('✅ Working!');
      } else {
        setResult(`Error: ${keys.error}`);
        setStatus('❌ Error');
      }
    } catch (err) {
      setResult(`Exception: ${err.message}`);
      setStatus('❌ Exception');
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Lighter SDK Test</Text>
      <Text style={styles.status}>Status: {status}</Text>
      
      <Button title="Test Key Generation" onPress={testGeneration} />
      
      <ScrollView style={styles.resultContainer}>
        <Text style={styles.result}>{result}</Text>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
    padding: 20,
    paddingTop: 60,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 20,
  },
  status: {
    fontSize: 16,
    marginBottom: 20,
  },
  resultContainer: {
    marginTop: 20,
    padding: 10,
    backgroundColor: '#f0f0f0',
    borderRadius: 5,
  },
  result: {
    fontFamily: 'Courier',
    fontSize: 12,
  },
});
```

## Troubleshooting

### Issue: "Module not found: expo-lighter-module"

**Solution:**
```bash
# Clean install
rm -rf node_modules
npm install

# Rebuild
npx expo prebuild --clean
```

### Issue: "Could not find module 'Lighter'"

**Solution:**
Verify the framework is in the right place:
```bash
ls -la expo-lighter-module/ios/Lighter.xcframework
```

If missing:
```bash
# From lighter-go root
./mobile/build_ios.sh
cp -r build/Lighter.xcframework expo-lighter-module/ios/
```

### Issue: Pod install fails

**Solution:**
```bash
cd ios
pod deintegrate
pod repo update
pod install
cd ..
```

### Issue: "Library not loaded: @rpath/Lighter.framework"

**Solution:**
The framework needs to be embedded. Check `ExpoLighterModule.podspec`:
```ruby
s.vendored_frameworks = "Lighter.xcframework"
```

Then:
```bash
cd ios
pod install
cd ..
```

## Building for Production

### 1. Update Build Number

In `app.json`:
```json
{
  "expo": {
    "version": "1.0.0",
    "ios": {
      "buildNumber": "1"
    }
  }
}
```

### 2. Build with EAS

```bash
# Production build
eas build --platform ios --profile production

# After build completes, submit to App Store
eas submit --platform ios
```

### 3. Manual Xcode Build

```bash
# Open in Xcode
open ios/YourApp.xcworkspace

# Then:
# 1. Select your team in Signing & Capabilities
# 2. Select a device or Any iOS Device
# 3. Product → Archive
# 4. Distribute App → App Store Connect
```

## Next Steps

- Implement secure key storage using `expo-secure-store`
- Add environment variables for API endpoints
- Implement error handling and logging
- Add unit tests for trading logic
- Set up continuous integration

## Security Checklist

- [ ] Store private keys in `expo-secure-store`, never in code
- [ ] Use environment variables for API URLs
- [ ] Implement biometric authentication
- [ ] Enable SSL pinning for API calls
- [ ] Validate all user inputs
- [ ] Implement rate limiting
- [ ] Add transaction confirmation flows
- [ ] Set up error monitoring (Sentry, etc.)

## Resources

- [Expo Documentation](https://docs.expo.dev/)
- [Expo Modules API](https://docs.expo.dev/modules/overview/)
- [EAS Build](https://docs.expo.dev/build/introduction/)
- [Lighter Documentation](https://docs.lighter.xyz/)

## Support

For issues specific to this module:
- GitHub: https://github.com/elliottech/lighter-go/issues

For Expo-related issues:
- Expo Forums: https://forums.expo.dev/
- Discord: https://chat.expo.dev/

