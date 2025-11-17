import { requireNativeModule } from 'expo-modules-core';

// Types
export interface APIKeyResult {
  privateKey: string;
  publicKey: string;
  error: string;
}

export interface TxResult {
  json: string;
  error: string;
}

// Native module
const ExpoLighter = requireNativeModule('ExpoLighter');

/**
 * Lighter SDK for React Native/Expo
 */
export class LighterSDK {
  // MARK: - Key Management

  /**
   * Generate a new API key pair
   * @param seed Optional seed string for deterministic key generation. Pass empty string for random.
   */
  static generateAPIKey(seed: string = ''): APIKeyResult {
    return ExpoLighter.generateAPIKey(seed);
  }

  /**
   * Create a new client for trading
   * @returns Empty string on success, error message on failure
   */
  static createClient(
    url: string,
    privateKey: string,
    chainId: number,
    apiKeyIndex: number,
    accountIndex: number
  ): string {
    return ExpoLighter.createClient(url, privateKey, chainId, apiKeyIndex, accountIndex);
  }

  /**
   * Check that the client is properly configured
   * @returns Empty string on success, error message on failure
   */
  static checkClient(apiKeyIndex: number, accountIndex: number): string {
    return ExpoLighter.checkClient(apiKeyIndex, accountIndex);
  }

  /**
   * Switch between multiple API keys
   * @returns Empty string on success, error message on failure
   */
  static switchAPIKey(apiKeyIndex: number): string {
    return ExpoLighter.switchAPIKey(apiKeyIndex);
  }

  // MARK: - Trading Operations

  /**
   * Sign a create order transaction
   * @param orderExpiry Unix timestamp in milliseconds, -1 for default (28 days)
   * @param nonce Transaction nonce, -1 for automatic
   */
  static signCreateOrder(params: {
    marketIndex: number;
    clientOrderIndex: number;
    baseAmount: number;
    price: number;
    isAsk: number; // 0 = buy, 1 = sell
    orderType: number; // 0 = limit, 1 = market, etc.
    timeInForce: number;
    reduceOnly: number;
    triggerPrice: number;
    orderExpiry?: number; // default -1
    nonce?: number; // default -1
  }): TxResult {
    return ExpoLighter.signCreateOrder({
      marketIndex: params.marketIndex,
      clientOrderIndex: params.clientOrderIndex,
      baseAmount: params.baseAmount,
      price: params.price,
      isAsk: params.isAsk,
      orderType: params.orderType,
      timeInForce: params.timeInForce,
      reduceOnly: params.reduceOnly,
      triggerPrice: params.triggerPrice,
      orderExpiry: params.orderExpiry ?? -1,
      nonce: params.nonce ?? -1
    });
  }

  /**
   * Sign a create grouped orders transaction (OCO, OTO, OTOCO)
   * @param groupingType 0 = None, 1 = OCO (One-Cancels-Other), 2 = OTO (One-Triggers-Other), 3 = OTOCO
   * @param orders Array of orders
   * @param nonce Transaction nonce, -1 for automatic
   */
  static signCreateGroupedOrders(
    groupingType: number,
    orders: Array<{
      marketIndex: number;
      clientOrderIndex: number;
      baseAmount: number;
      price: number;
      isAsk: number;
      type: number;
      timeInForce: number;
      reduceOnly?: number;
      triggerPrice?: number;
      orderExpiry?: number;
    }>,
    nonce: number = -1
  ): TxResult {
    const ordersJSON = JSON.stringify(orders);
    return ExpoLighter.signCreateGroupedOrders(groupingType, ordersJSON, nonce);
  }

  /**
   * Sign a cancel order transaction
   */
  static signCancelOrder(
    marketIndex: number,
    orderIndex: number,
    nonce: number = -1
  ): TxResult {
    return ExpoLighter.signCancelOrder(marketIndex, orderIndex, nonce);
  }

  /**
   * Sign a cancel all orders transaction
   */
  static signCancelAllOrders(
    timeInForce: number,
    time: number,
    nonce: number = -1
  ): TxResult {
    return ExpoLighter.signCancelAllOrders(timeInForce, time, nonce);
  }

  /**
   * Sign a modify order transaction
   */
  static signModifyOrder(params: {
    marketIndex: number;
    index: number;
    baseAmount: number;
    price: number;
    triggerPrice: number;
    nonce?: number;
  }): TxResult {
    return ExpoLighter.signModifyOrder(
      params.marketIndex,
      params.index,
      params.baseAmount,
      params.price,
      params.triggerPrice,
      params.nonce ?? -1
    );
  }

  // MARK: - Account Management

  /**
   * Sign a withdraw transaction
   * @param usdcAmount Amount in USDC base units
   */
  static signWithdraw(usdcAmount: number, nonce: number = -1): TxResult {
    return ExpoLighter.signWithdraw(usdcAmount, nonce);
  }

  /**
   * Sign a transfer transaction
   * @param memo Must be exactly 32 bytes
   */
  static signTransfer(params: {
    toAccountIndex: number;
    usdcAmount: number;
    fee: number;
    memo: string; // Must be 32 bytes
    nonce?: number;
  }): TxResult {
    return ExpoLighter.signTransfer(
      params.toAccountIndex,
      params.usdcAmount,
      params.fee,
      params.memo,
      params.nonce ?? -1
    );
  }

  /**
   * Sign a create sub-account transaction
   */
  static signCreateSubAccount(nonce: number = -1): TxResult {
    return ExpoLighter.signCreateSubAccount(nonce);
  }

  /**
   * Sign a change public key transaction
   * @param pubKey Hex-encoded public key (40 bytes)
   */
  static signChangePubKey(pubKey: string, nonce: number = -1): TxResult {
    return ExpoLighter.signChangePubKey(pubKey, nonce);
  }

  // MARK: - Pool Operations

  /**
   * Sign a create public pool transaction
   */
  static signCreatePublicPool(params: {
    operatorFee: number;
    initialTotalShares: number;
    minOperatorShareRate: number;
    nonce?: number;
  }): TxResult {
    return ExpoLighter.signCreatePublicPool(
      params.operatorFee,
      params.initialTotalShares,
      params.minOperatorShareRate,
      params.nonce ?? -1
    );
  }

  /**
   * Sign an update public pool transaction
   */
  static signUpdatePublicPool(params: {
    publicPoolIndex: number;
    status: number;
    operatorFee: number;
    minOperatorShareRate: number;
    nonce?: number;
  }): TxResult {
    return ExpoLighter.signUpdatePublicPool(
      params.publicPoolIndex,
      params.status,
      params.operatorFee,
      params.minOperatorShareRate,
      params.nonce ?? -1
    );
  }

  /**
   * Sign a mint shares transaction
   */
  static signMintShares(
    publicPoolIndex: number,
    shareAmount: number,
    nonce: number = -1
  ): TxResult {
    return ExpoLighter.signMintShares(publicPoolIndex, shareAmount, nonce);
  }

  /**
   * Sign a burn shares transaction
   */
  static signBurnShares(
    publicPoolIndex: number,
    shareAmount: number,
    nonce: number = -1
  ): TxResult {
    return ExpoLighter.signBurnShares(publicPoolIndex, shareAmount, nonce);
  }

  // MARK: - Position Management

  /**
   * Sign an update leverage transaction
   */
  static signUpdateLeverage(params: {
    marketIndex: number;
    initialMarginFraction: number;
    marginMode: number;
    nonce?: number;
  }): TxResult {
    return ExpoLighter.signUpdateLeverage(
      params.marketIndex,
      params.initialMarginFraction,
      params.marginMode,
      params.nonce ?? -1
    );
  }

  /**
   * Sign an update margin transaction
   */
  static signUpdateMargin(params: {
    marketIndex: number;
    usdcAmount: number;
    direction: number;
    nonce?: number;
  }): TxResult {
    return ExpoLighter.signUpdateMargin(
      params.marketIndex,
      params.usdcAmount,
      params.direction,
      params.nonce ?? -1
    );
  }

  // MARK: - Authentication

  /**
   * Create an authentication token
   * @param deadline Unix timestamp, 0 for default (7 hours from now)
   */
  static createAuthToken(deadline: number = 0): TxResult {
    return ExpoLighter.createAuthToken(deadline);
  }
}

// Export everything
export default LighterSDK;

