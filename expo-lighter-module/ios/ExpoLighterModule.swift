import ExpoModulesCore
import Lighter

public class ExpoLighterModule: Module {
  public func definition() -> ModuleDefinition {
    Name("ExpoLighter")

    // MARK: - Key Management
    
    Function("generateAPIKey") { (seed: String?) -> [String: Any] in
      let result = MobileGenerateAPIKey(seed ?? "")
      return [
        "privateKey": result?.privateKey ?? "",
        "publicKey": result?.publicKey ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("createClient") { (url: String, privateKey: String, chainId: Int, apiKeyIndex: Int, accountIndex: Int64) -> String in
      return MobileCreateClient(url, privateKey, chainId, apiKeyIndex, accountIndex)
    }

    Function("checkClient") { (apiKeyIndex: Int, accountIndex: Int64) -> String in
      return MobileCheckClient(apiKeyIndex, accountIndex)
    }

    Function("switchAPIKey") { (apiKeyIndex: Int) -> String in
      return MobileSwitchAPIKey(apiKeyIndex)
    }

    // MARK: - Trading Operations

    Function("signCreateOrder") { (params: [String: Any]) -> [String: String] in
      guard let marketIndex = params["marketIndex"] as? Int,
            let clientOrderIndex = params["clientOrderIndex"] as? Int64,
            let baseAmount = params["baseAmount"] as? Int64,
            let price = params["price"] as? Int,
            let isAsk = params["isAsk"] as? Int,
            let orderType = params["orderType"] as? Int,
            let timeInForce = params["timeInForce"] as? Int,
            let reduceOnly = params["reduceOnly"] as? Int,
            let triggerPrice = params["triggerPrice"] as? Int else {
        return [
          "json": "",
          "error": "Missing required parameters for signCreateOrder"
        ]
      }
      
      let orderExpiry = params["orderExpiry"] as? Int64 ?? -1
      let nonce = params["nonce"] as? Int64 ?? -1
      
      let result = MobileSignCreateOrder(
        marketIndex,
        clientOrderIndex,
        baseAmount,
        price,
        isAsk,
        orderType,
        timeInForce,
        reduceOnly,
        triggerPrice,
        orderExpiry,
        nonce
      )
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signCreateGroupedOrders") { (groupingType: Int, ordersJSON: String, nonce: Int64) -> [String: String] in
      let result = MobileSignCreateGroupedOrders(groupingType, ordersJSON, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signCancelOrder") { (marketIndex: Int, orderIndex: Int64, nonce: Int64) -> [String: String] in
      let result = MobileSignCancelOrder(marketIndex, orderIndex, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signCancelAllOrders") { (timeInForce: Int, time: Int64, nonce: Int64) -> [String: String] in
      let result = MobileSignCancelAllOrders(timeInForce, time, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signModifyOrder") { (
      marketIndex: Int,
      index: Int64,
      baseAmount: Int64,
      price: Int64,
      triggerPrice: Int64,
      nonce: Int64
    ) -> [String: String] in
      let result = MobileSignModifyOrder(
        marketIndex,
        index,
        baseAmount,
        price,
        triggerPrice,
        nonce
      )
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    // MARK: - Account Management

    Function("signWithdraw") { (usdcAmount: Int64, nonce: Int64) -> [String: String] in
      let result = MobileSignWithdraw(usdcAmount, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signTransfer") { (
      toAccountIndex: Int64,
      usdcAmount: Int64,
      fee: Int64,
      memo: String,
      nonce: Int64
    ) -> [String: String] in
      let result = MobileSignTransfer(
        toAccountIndex,
        usdcAmount,
        fee,
        memo,
        nonce
      )
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signCreateSubAccount") { (nonce: Int64) -> [String: String] in
      let result = MobileSignCreateSubAccount(nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signChangePubKey") { (pubKey: String, nonce: Int64) -> [String: String] in
      let result = MobileSignChangePubKey(pubKey, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    // MARK: - Pool Operations

    Function("signCreatePublicPool") { (
      operatorFee: Int64,
      initialTotalShares: Int64,
      minOperatorShareRate: Int64,
      nonce: Int64
    ) -> [String: String] in
      let result = MobileSignCreatePublicPool(
        operatorFee,
        initialTotalShares,
        minOperatorShareRate,
        nonce
      )
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signUpdatePublicPool") { (
      publicPoolIndex: Int64,
      status: Int,
      operatorFee: Int64,
      minOperatorShareRate: Int64,
      nonce: Int64
    ) -> [String: String] in
      let result = MobileSignUpdatePublicPool(
        publicPoolIndex,
        status,
        operatorFee,
        minOperatorShareRate,
        nonce
      )
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signMintShares") { (publicPoolIndex: Int64, shareAmount: Int64, nonce: Int64) -> [String: String] in
      let result = MobileSignMintShares(publicPoolIndex, shareAmount, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signBurnShares") { (publicPoolIndex: Int64, shareAmount: Int64, nonce: Int64) -> [String: String] in
      let result = MobileSignBurnShares(publicPoolIndex, shareAmount, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    // MARK: - Position Management

    Function("signUpdateLeverage") { (
      marketIndex: Int,
      initialMarginFraction: Int,
      marginMode: Int,
      nonce: Int64
    ) -> [String: String] in
      let result = MobileSignUpdateLeverage(marketIndex, initialMarginFraction, marginMode, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    Function("signUpdateMargin") { (
      marketIndex: Int,
      usdcAmount: Int64,
      direction: Int,
      nonce: Int64
    ) -> [String: String] in
      let result = MobileSignUpdateMargin(marketIndex, usdcAmount, direction, nonce)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }

    // MARK: - Authentication

    Function("createAuthToken") { (deadline: Int64) -> [String: String] in
      let result = MobileCreateAuthToken(deadline)
      return [
        "json": result?.json ?? "",
        "error": result?.error ?? ""
      ]
    }
  }
}

