//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"
	"time"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types"
	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	txClient        *client.TxClient
	backupTxClients map[uint8]*client.TxClient
)

func generateAPIKey(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in generateAPIKey: %v\n", r)
		}
	}()

	seed := ""
	if len(args) > 0 && args[0].Type() == js.TypeString {
		seed = args[0].String()
	}

	var seedP *string
	if seed != "" {
		seedP = &seed
	}

	key := curve.SampleScalar(seedP)
	publicKeyStr := hexutil.Encode(schnorr.SchnorrPkFromSk(key).ToLittleEndianBytes())
	privateKeyStr := hexutil.Encode(key.ToLittleEndianBytes())

	result := map[string]interface{}{
		"privateKey": privateKeyStr,
		"publicKey":  publicKeyStr,
		"err":        nil,
	}

	return js.ValueOf(result)
}

func createClient(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in createClient: %v\n", r)
		}
	}()

	if len(args) < 5 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	url := args[0].String()
	privateKey := args[1].String()
	chainId := uint32(args[2].Int())
	apiKeyIndex := uint8(args[3].Int())
	accountIndex := int64(args[4].Int())

	if accountIndex <= 0 {
		return js.ValueOf(map[string]interface{}{
			"err": "invalid account index",
		})
	}

	httpClient := client.NewHTTPClient(url)
	var err error
	txClient, err = client.NewTxClient(httpClient, privateKey, accountIndex, apiKeyIndex, chainId)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	if backupTxClients == nil {
		backupTxClients = make(map[uint8]*client.TxClient)
	}
	backupTxClients[apiKeyIndex] = txClient

	return js.ValueOf(map[string]interface{}{
		"err": nil,
	})
}

func checkClient(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in checkClient: %v\n", r)
		}
	}()

	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	apiKeyIndex := uint8(args[0].Int())
	accountIndex := int64(args[1].Int())

	client, ok := backupTxClients[apiKeyIndex]
	if !ok {
		return js.ValueOf(map[string]interface{}{
			"err": "api key not registered",
		})
	}

	if client.GetApiKeyIndex() != apiKeyIndex {
		return js.ValueOf(map[string]interface{}{
			"err": fmt.Sprintf("apiKeyIndex does not match. expected %v but got %v", client.GetApiKeyIndex(), apiKeyIndex),
		})
	}
	if client.GetAccountIndex() != accountIndex {
		return js.ValueOf(map[string]interface{}{
			"err": fmt.Sprintf("accountIndex does not match. expected %v but got %v", client.GetAccountIndex(), accountIndex),
		})
	}

	// check that the API key registered on Lighter matches this one
	key, err := client.HTTP().GetApiKey(accountIndex, apiKeyIndex)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": fmt.Sprintf("failed to get Api Keys. err: %v", err),
		})
	}

	pubKeyBytes := client.GetKeyManager().PubKeyBytes()
	pubKeyStr := hexutil.Encode(pubKeyBytes[:])
	pubKeyStr = strings.Replace(pubKeyStr, "0x", "", 1)

	ak := key.ApiKeys[0]
	if ak.PublicKey != pubKeyStr {
		return js.ValueOf(map[string]interface{}{
			"err": fmt.Sprintf("private key does not match the one on Lighter. ownPubKey: %s response: %+v", pubKeyStr, ak),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"err": nil,
	})
}

func signChangePubKey(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signChangePubKey: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	pubKeyStr := args[0].String()
	nonce := int64(args[1].Int())

	// handle PubKey
	pubKeyBytes, err := hexutil.Decode(pubKeyStr)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}
	if len(pubKeyBytes) != 40 {
		return js.ValueOf(map[string]interface{}{
			"err": fmt.Sprintf("invalid pub key length. expected 40 but got %v", len(pubKeyBytes)),
		})
	}
	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	txInfo := &types.ChangePubKeyReq{
		PubKey: pubKey,
	}
	ops := &types.TransactOpts{}
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetChangePubKeyTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	// === manually add MessageToSign to the response:
	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}
	obj := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &obj)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}
	obj["MessageToSign"] = tx.GetL1SignatureBody()
	txInfoBytes, err = json.Marshal(obj)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signCreateOrder(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signCreateOrder: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 11 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	marketIndex := uint8(args[0].Int())
	clientOrderIndex := int64(args[1].Int())
	baseAmount := int64(args[2].Int())
	price := uint32(args[3].Int())
	isAsk := uint8(args[4].Int())
	orderType := uint8(args[5].Int())
	timeInForce := uint8(args[6].Int())
	reduceOnly := uint8(args[7].Int())
	triggerPrice := uint32(args[8].Int())
	orderExpiry := int64(args[9].Int())
	nonce := int64(args[10].Int())

	if orderExpiry == -1 {
		orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli() // 28 days
	}

	txInfo := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: clientOrderIndex,
		BaseAmount:       baseAmount,
		Price:            price,
		IsAsk:            isAsk,
		Type:             orderType,
		TimeInForce:      timeInForce,
		ReduceOnly:       reduceOnly,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      orderExpiry,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreateOrderTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signCancelOrder(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signCancelOrder: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 3 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	marketIndex := uint8(args[0].Int())
	orderIndex := int64(args[1].Int())
	nonce := int64(args[2].Int())

	txInfo := &types.CancelOrderTxReq{
		MarketIndex: marketIndex,
		Index:       orderIndex,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCancelOrderTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signCreateGroupedOrders(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signCreateGroupedOrders: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 3 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments: expected groupingType, ordersJSON, nonce",
		})
	}

	groupingType := uint8(args[0].Int())
	ordersJSON := args[1].String()
	nonce := int64(args[2].Int())

	// Parse orders from JSON
	var ordersData []map[string]interface{}
	if err := json.Unmarshal([]byte(ordersJSON), &ordersData); err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": fmt.Sprintf("failed to parse orders JSON: %v", err),
		})
	}

	// Convert to CreateOrderTxReq slice
	orders := make([]*types.CreateOrderTxReq, 0, len(ordersData))
	for i, orderData := range ordersData {
		marketIndex, ok := orderData["marketIndex"].(float64)
		if !ok {
			return js.ValueOf(map[string]interface{}{
				"err": fmt.Sprintf("order %d: marketIndex is required", i),
			})
		}

		clientOrderIndex, ok := orderData["clientOrderIndex"].(float64)
		if !ok {
			return js.ValueOf(map[string]interface{}{
				"err": fmt.Sprintf("order %d: clientOrderIndex is required", i),
			})
		}

		baseAmount, ok := orderData["baseAmount"].(float64)
		if !ok {
			return js.ValueOf(map[string]interface{}{
				"err": fmt.Sprintf("order %d: baseAmount is required", i),
			})
		}

		price, ok := orderData["price"].(float64)
		if !ok {
			return js.ValueOf(map[string]interface{}{
				"err": fmt.Sprintf("order %d: price is required", i),
			})
		}

		isAsk, ok := orderData["isAsk"].(float64)
		if !ok {
			return js.ValueOf(map[string]interface{}{
				"err": fmt.Sprintf("order %d: isAsk is required", i),
			})
		}

		orderType, ok := orderData["type"].(float64)
		if !ok {
			return js.ValueOf(map[string]interface{}{
				"err": fmt.Sprintf("order %d: type is required", i),
			})
		}

		timeInForce, ok := orderData["timeInForce"].(float64)
		if !ok {
			return js.ValueOf(map[string]interface{}{
				"err": fmt.Sprintf("order %d: timeInForce is required", i),
			})
		}

		reduceOnly := uint8(0)
		if ro, ok := orderData["reduceOnly"].(float64); ok {
			reduceOnly = uint8(ro)
		}

		triggerPrice := uint32(0)
		if tp, ok := orderData["triggerPrice"].(float64); ok {
			triggerPrice = uint32(tp)
		}

		orderExpiry := int64(-1)
		if oe, ok := orderData["orderExpiry"].(float64); ok {
			orderExpiry = int64(oe)
		}
		if orderExpiry == -1 {
			orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli() // 28 days
		}

		orders = append(orders, &types.CreateOrderTxReq{
			MarketIndex:      uint8(marketIndex),
			ClientOrderIndex: int64(clientOrderIndex),
			BaseAmount:       int64(baseAmount),
			Price:            uint32(price),
			IsAsk:            uint8(isAsk),
			Type:             uint8(orderType),
			TimeInForce:      uint8(timeInForce),
			ReduceOnly:       reduceOnly,
			TriggerPrice:     triggerPrice,
			OrderExpiry:      orderExpiry,
		})
	}

	txInfo := &types.CreateGroupedOrdersTxReq{
		GroupingType: groupingType,
		Orders:       orders,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreateGroupedOrdersTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signWithdraw(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signWithdraw: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	usdcAmount := uint64(args[0].Int())
	nonce := int64(args[1].Int())

	txInfo := types.WithdrawTxReq{
		USDCAmount: usdcAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetWithdrawTransaction(&txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signCreateSubAccount(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signCreateSubAccount: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	nonce := int64(args[0].Int())

	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreateSubAccountTransaction(ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signCancelAllOrders(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signCancelAllOrders: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 3 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	timeInForce := uint8(args[0].Int())
	t := int64(args[1].Int())
	nonce := int64(args[2].Int())

	txInfo := &types.CancelAllOrdersTxReq{
		TimeInForce: timeInForce,
		Time:        t,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCancelAllOrdersTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signModifyOrder(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signModifyOrder: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 6 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	marketIndex := uint8(args[0].Int())
	index := int64(args[1].Int())
	baseAmount := int64(args[2].Int())
	price := uint32(args[3].Int())
	triggerPrice := uint32(args[4].Int())
	nonce := int64(args[5].Int())

	txInfo := &types.ModifyOrderTxReq{
		MarketIndex:  marketIndex,
		Index:        index,
		BaseAmount:   baseAmount,
		Price:        price,
		TriggerPrice: triggerPrice,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetModifyOrderTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signTransfer(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signTransfer: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 5 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	toAccountIndex := int64(args[0].Int())
	usdcAmount := int64(args[1].Int())
	fee := int64(args[2].Int())
	memoStr := args[3].String()
	nonce := int64(args[4].Int())

	memo := [32]byte{}
	if len(memoStr) != 32 {
		return js.ValueOf(map[string]interface{}{
			"err": "memo expected to be 32 bytes long",
		})
	}
	for i := 0; i < 32; i++ {
		memo[i] = byte(memoStr[i])
	}

	txInfo := &types.TransferTxReq{
		ToAccountIndex: toAccountIndex,
		USDCAmount:     usdcAmount,
		Fee:            fee,
		Memo:           memo,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetTransferTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}
	
	// Add MessageToSign like in original
	obj := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &obj)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}
	obj["MessageToSign"] = tx.GetL1SignatureBody()
	txInfoBytes, err = json.Marshal(obj)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signCreatePublicPool(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signCreatePublicPool: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 4 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	operatorFee := int64(args[0].Int())
	initialTotalShares := int64(args[1].Int())
	minOperatorShareRate := int64(args[2].Int())
	nonce := int64(args[3].Int())

	txInfo := &types.CreatePublicPoolTxReq{
		OperatorFee:          operatorFee,
		InitialTotalShares:   initialTotalShares,
		MinOperatorShareRate: minOperatorShareRate,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreatePublicPoolTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signUpdatePublicPool(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signUpdatePublicPool: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 5 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	publicPoolIndex := int64(args[0].Int())
	status := uint8(args[1].Int())
	operatorFee := int64(args[2].Int())
	minOperatorShareRate := int64(args[3].Int())
	nonce := int64(args[4].Int())

	txInfo := &types.UpdatePublicPoolTxReq{
		PublicPoolIndex:      publicPoolIndex,
		Status:               status,
		OperatorFee:          operatorFee,
		MinOperatorShareRate: minOperatorShareRate,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdatePublicPoolTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signMintShares(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signMintShares: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 3 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	publicPoolIndex := int64(args[0].Int())
	shareAmount := int64(args[1].Int())
	nonce := int64(args[2].Int())

	txInfo := &types.MintSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetMintSharesTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signBurnShares(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signBurnShares: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 3 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	publicPoolIndex := int64(args[0].Int())
	shareAmount := int64(args[1].Int())
	nonce := int64(args[2].Int())

	txInfo := &types.BurnSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetBurnSharesTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func signUpdateLeverage(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signUpdateLeverage: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 4 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	marketIndex := uint8(args[0].Int())
	initialMarginFraction := uint16(args[1].Int())
	marginMode := uint8(args[2].Int())
	nonce := int64(args[3].Int())

	txInfo := &types.UpdateLeverageTxReq{
		MarketIndex:           marketIndex,
		InitialMarginFraction: initialMarginFraction,
		MarginMode:            marginMode,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdateLeverageTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func createAuthToken(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in createAuthToken: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "client is not created, call createClient() first",
		})
	}

	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	deadline := int64(args[0].Int())
	if deadline == 0 {
		deadline = time.Now().Add(time.Hour * 7).Unix()
	}

	authToken, err := txClient.GetAuthToken(time.Unix(deadline, 0))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": authToken,
		"err": nil,
	})
}

func switchAPIKey(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in switchAPIKey: %v\n", r)
		}
	}()

	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	apiKeyIndex := uint8(args[0].Int())

	client, ok := backupTxClients[apiKeyIndex]
	if !ok {
		return js.ValueOf(map[string]interface{}{
			"err": "api key not registered",
		})
	}

	txClient = client

	return js.ValueOf(map[string]interface{}{
		"err": nil,
	})
}

func signUpdateMargin(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in signUpdateMargin: %v\n", r)
		}
	}()

	if txClient == nil {
		return js.ValueOf(map[string]interface{}{
			"err": "Client is not created, call CreateClient() first",
		})
	}

	if len(args) < 4 {
		return js.ValueOf(map[string]interface{}{
			"err": "insufficient arguments",
		})
	}

	marketIndex := uint8(args[0].Int())
	usdcAmount := int64(args[1].Int())
	direction := uint8(args[2].Int())
	nonce := int64(args[3].Int())

	txInfo := &types.UpdateMarginTxReq{
		MarketIndex: marketIndex,
		USDCAmount:  usdcAmount,
		Direction:   direction,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdateMarginTransaction(txInfo, ops)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"err": err.Error(),
		})
	}

	return js.ValueOf(map[string]interface{}{
		"str": string(txInfoBytes),
		"err": nil,
	})
}

func main() {
	// Register functions to be called from JavaScript
	js.Global().Set("GenerateAPIKey", js.FuncOf(generateAPIKey))
	js.Global().Set("CreateClient", js.FuncOf(createClient))
	js.Global().Set("CheckClient", js.FuncOf(checkClient))
	js.Global().Set("SignChangePubKey", js.FuncOf(signChangePubKey))
	js.Global().Set("SignCreateOrder", js.FuncOf(signCreateOrder))
	js.Global().Set("SignCancelOrder", js.FuncOf(signCancelOrder))
	js.Global().Set("SignCreateGroupedOrders", js.FuncOf(signCreateGroupedOrders))
	js.Global().Set("SignWithdraw", js.FuncOf(signWithdraw))
	js.Global().Set("SignCreateSubAccount", js.FuncOf(signCreateSubAccount))
	js.Global().Set("SignCancelAllOrders", js.FuncOf(signCancelAllOrders))
	js.Global().Set("SignModifyOrder", js.FuncOf(signModifyOrder))
	js.Global().Set("SignTransfer", js.FuncOf(signTransfer))
	js.Global().Set("SignCreatePublicPool", js.FuncOf(signCreatePublicPool))
	js.Global().Set("SignUpdatePublicPool", js.FuncOf(signUpdatePublicPool))
	js.Global().Set("SignMintShares", js.FuncOf(signMintShares))
	js.Global().Set("SignBurnShares", js.FuncOf(signBurnShares))
	js.Global().Set("SignUpdateLeverage", js.FuncOf(signUpdateLeverage))
	js.Global().Set("CreateAuthToken", js.FuncOf(createAuthToken))
	js.Global().Set("SwitchAPIKey", js.FuncOf(switchAPIKey))
	js.Global().Set("SignUpdateMargin", js.FuncOf(signUpdateMargin))

	fmt.Println("Lighter Go WASM module loaded successfully")

	// Keep the program running
	select {}
}
