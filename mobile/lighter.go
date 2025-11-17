package mobile

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types"
	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	txClient        *client.TxClient
	backupTxClients map[int]*client.TxClient
)

// GenerateAPIKey generates a new API key pair from an optional seed
// Pass empty string for seed to generate random key
func GenerateAPIKey(seed string) *APIKeyResult {
	defer func() {
		if r := recover(); r != nil {
			// This shouldn't happen, but just in case
		}
	}()

	seedP := &seed
	if seed == "" {
		seedP = nil
	}

	key := curve.SampleScalar(seedP)
	publicKeyStr := hexutil.Encode(schnorr.SchnorrPkFromSk(key).ToLittleEndianBytes())
	privateKeyStr := hexutil.Encode(key.ToLittleEndianBytes())

	return &APIKeyResult{
		PrivateKey: privateKeyStr,
		PublicKey:  publicKeyStr,
		Error:      "",
	}
}

// CreateClient creates a new transaction client
// url: API endpoint URL
// privateKey: hex-encoded private key
// chainId: blockchain chain ID
// apiKeyIndex: index of the API key (0-255)
// accountIndex: account index (must be > 0)
func CreateClient(url, privateKey string, chainId, apiKeyIndex int, accountIndex int64) string {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if accountIndex <= 0 {
		return "invalid account index"
	}

	httpClient := client.NewHTTPClient(url)
	var err error
	txClient, err = client.NewTxClient(httpClient, privateKey, accountIndex, uint8(apiKeyIndex), uint32(chainId))
	if err != nil {
		return fmt.Sprintf("error occurred when creating TxClient. err: %v", err)
	}

	if backupTxClients == nil {
		backupTxClients = make(map[int]*client.TxClient)
	}
	backupTxClients[apiKeyIndex] = txClient

	return ""
}

// CheckClient verifies that the client is properly configured and matches the API key on Lighter
func CheckClient(apiKeyIndex int, accountIndex int64) string {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	client, ok := backupTxClients[apiKeyIndex]
	if !ok {
		return "api key not registered"
	}

	if client.GetApiKeyIndex() != uint8(apiKeyIndex) {
		return fmt.Sprintf("apiKeyIndex does not match. expected %v but got %v", client.GetApiKeyIndex(), apiKeyIndex)
	}
	if client.GetAccountIndex() != accountIndex {
		return fmt.Sprintf("accountIndex does not match. expected %v but got %v", client.GetAccountIndex(), accountIndex)
	}

	// check that the API key registered on Lighter matches this one
	key, err := client.HTTP().GetApiKey(accountIndex, uint8(apiKeyIndex))
	if err != nil {
		return fmt.Sprintf("failed to get Api Keys. err: %v", err)
	}

	pubKeyBytes := client.GetKeyManager().PubKeyBytes()
	pubKeyStr := hexutil.Encode(pubKeyBytes[:])
	pubKeyStr = strings.Replace(pubKeyStr, "0x", "", 1)

	ak := key.ApiKeys[0]
	if ak.PublicKey != pubKeyStr {
		return fmt.Sprintf("private key does not match the one on Lighter. ownPubKey: %s response: %+v", pubKeyStr, ak)
	}

	return ""
}

// SignChangePubKey signs a change public key transaction
// pubKey: hex-encoded new public key (40 bytes)
// nonce: transaction nonce, use -1 for automatic
func SignChangePubKey(pubKey string, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	// handle PubKey
	pubKeyBytes, err := hexutil.Decode(pubKey)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}
	if len(pubKeyBytes) != 40 {
		return &TxResult{Error: fmt.Sprintf("invalid pub key length. expected 40 but got %v", len(pubKeyBytes))}
	}
	var pubKeyArray [40]byte
	copy(pubKeyArray[:], pubKeyBytes)

	txInfo := &types.ChangePubKeyReq{
		PubKey: pubKeyArray,
	}
	ops := &types.TransactOpts{}
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetChangePubKeyTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	// marshal the tx
	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}
	obj := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &obj)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}
	obj["MessageToSign"] = tx.GetL1SignatureBody()
	txInfoBytes, err = json.Marshal(obj)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignCreateOrder signs a create order transaction
// Pass -1 for orderExpiry to use default (28 days)
// Pass -1 for nonce for automatic nonce
func SignCreateOrder(marketIndex int, clientOrderIndex, baseAmount int64, price int,
	isAsk, orderType, timeInForce, reduceOnly int, triggerPrice int,
	orderExpiry, nonce int64) *TxResult {

	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	if orderExpiry == -1 {
		orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli() // 28 days
	}

	txInfo := &types.CreateOrderTxReq{
		MarketIndex:      uint8(marketIndex),
		ClientOrderIndex: clientOrderIndex,
		BaseAmount:       baseAmount,
		Price:            uint32(price),
		IsAsk:            uint8(isAsk),
		Type:             uint8(orderType),
		TimeInForce:      uint8(timeInForce),
		ReduceOnly:       uint8(reduceOnly),
		TriggerPrice:     uint32(triggerPrice),
		OrderExpiry:      orderExpiry,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreateOrderTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignCreateGroupedOrders signs a create grouped orders transaction
// ordersJSON should be a JSON array of order objects
// groupingType: 0 = None, 1 = OCO (One-Cancels-Other), 2 = OTO (One-Triggers-Other), 3 = OTOCO
func SignCreateGroupedOrders(groupingType int, ordersJSON string, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	// Parse orders from JSON
	var ordersData []map[string]interface{}
	if err := json.Unmarshal([]byte(ordersJSON), &ordersData); err != nil {
		return &TxResult{Error: fmt.Sprintf("failed to parse orders JSON: %v", err)}
	}

	// Convert to CreateOrderTxReq slice
	orders := make([]*types.CreateOrderTxReq, 0, len(ordersData))
	for i, orderData := range ordersData {
		marketIndex, ok := orderData["marketIndex"].(float64)
		if !ok {
			return &TxResult{Error: fmt.Sprintf("order %d: marketIndex is required", i)}
		}

		clientOrderIndex, ok := orderData["clientOrderIndex"].(float64)
		if !ok {
			return &TxResult{Error: fmt.Sprintf("order %d: clientOrderIndex is required", i)}
		}

		baseAmount, ok := orderData["baseAmount"].(float64)
		if !ok {
			return &TxResult{Error: fmt.Sprintf("order %d: baseAmount is required", i)}
		}

		price, ok := orderData["price"].(float64)
		if !ok {
			return &TxResult{Error: fmt.Sprintf("order %d: price is required", i)}
		}

		isAsk, ok := orderData["isAsk"].(float64)
		if !ok {
			return &TxResult{Error: fmt.Sprintf("order %d: isAsk is required", i)}
		}

		orderType, ok := orderData["type"].(float64)
		if !ok {
			return &TxResult{Error: fmt.Sprintf("order %d: type is required", i)}
		}

		timeInForce, ok := orderData["timeInForce"].(float64)
		if !ok {
			return &TxResult{Error: fmt.Sprintf("order %d: timeInForce is required", i)}
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
		GroupingType: uint8(groupingType),
		Orders:       orders,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreateGroupedOrdersTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignCancelOrder signs a cancel order transaction
func SignCancelOrder(marketIndex int, orderIndex, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	txInfo := &types.CancelOrderTxReq{
		MarketIndex: uint8(marketIndex),
		Index:       orderIndex,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCancelOrderTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignWithdraw signs a withdraw transaction
func SignWithdraw(usdcAmount int64, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	txInfo := types.WithdrawTxReq{
		USDCAmount: uint64(usdcAmount),
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetWithdrawTransaction(&txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignCreateSubAccount signs a create sub account transaction
func SignCreateSubAccount(nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreateSubAccountTransaction(ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignCancelAllOrders signs a cancel all orders transaction
func SignCancelAllOrders(timeInForce int, time, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	txInfo := &types.CancelAllOrdersTxReq{
		TimeInForce: uint8(timeInForce),
		Time:        time,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCancelAllOrdersTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignModifyOrder signs a modify order transaction
func SignModifyOrder(marketIndex int, index, baseAmount int64, price, triggerPrice int64, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	txInfo := &types.ModifyOrderTxReq{
		MarketIndex:  uint8(marketIndex),
		Index:        index,
		BaseAmount:   baseAmount,
		Price:        uint32(price),
		TriggerPrice: uint32(triggerPrice),
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetModifyOrderTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignTransfer signs a transfer transaction
// memo must be exactly 32 bytes
func SignTransfer(toAccountIndex, usdcAmount, fee int64, memo string, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	memoArray := [32]byte{}
	if len(memo) != 32 {
		return &TxResult{Error: "memo expected to be 32 bytes long"}
	}
	for i := 0; i < 32; i++ {
		memoArray[i] = byte(memo[i])
	}

	txInfo := &types.TransferTxReq{
		ToAccountIndex: toAccountIndex,
		USDCAmount:     usdcAmount,
		Fee:            fee,
		Memo:           memoArray,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetTransferTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}
	obj := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &obj)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}
	obj["MessageToSign"] = tx.GetL1SignatureBody()
	txInfoBytes, err = json.Marshal(obj)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignCreatePublicPool signs a create public pool transaction
func SignCreatePublicPool(operatorFee, initialTotalShares, minOperatorShareRate, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

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
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignUpdatePublicPool signs an update public pool transaction
func SignUpdatePublicPool(publicPoolIndex int64, status int, operatorFee, minOperatorShareRate, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	txInfo := &types.UpdatePublicPoolTxReq{
		PublicPoolIndex:      publicPoolIndex,
		Status:               uint8(status),
		OperatorFee:          operatorFee,
		MinOperatorShareRate: minOperatorShareRate,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdatePublicPoolTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignMintShares signs a mint shares transaction
func SignMintShares(publicPoolIndex, shareAmount, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

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
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignBurnShares signs a burn shares transaction
func SignBurnShares(publicPoolIndex, shareAmount, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

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
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignUpdateLeverage signs an update leverage transaction
func SignUpdateLeverage(marketIndex, initialMarginFraction, marginMode int, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	txInfo := &types.UpdateLeverageTxReq{
		MarketIndex:           uint8(marketIndex),
		InitialMarginFraction: uint16(initialMarginFraction),
		MarginMode:            uint8(marginMode),
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdateLeverageTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// SignUpdateMargin signs an update margin transaction
func SignUpdateMargin(marketIndex int, usdcAmount int64, direction int, nonce int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "Client is not created, call CreateClient() first"}
	}

	txInfo := &types.UpdateMarginTxReq{
		MarketIndex: uint8(marketIndex),
		USDCAmount:  usdcAmount,
		Direction:   uint8(direction),
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdateMarginTransaction(txInfo, ops)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: string(txInfoBytes), Error: ""}
}

// CreateAuthToken creates an authentication token
// deadline: Unix timestamp, use 0 for default (7 hours from now)
func CreateAuthToken(deadline int64) *TxResult {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	if txClient == nil {
		return &TxResult{Error: "client is not created, call CreateClient() first"}
	}

	if deadline == 0 {
		deadline = time.Now().Add(time.Hour * 7).Unix()
	}

	authToken, err := txClient.GetAuthToken(time.Unix(deadline, 0))
	if err != nil {
		return &TxResult{Error: err.Error()}
	}

	return &TxResult{JSON: authToken, Error: ""}
}

// SwitchAPIKey switches the active API key to the one at the specified index
func SwitchAPIKey(apiKeyIndex int) string {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	txClient = backupTxClients[apiKeyIndex]
	if txClient == nil {
		return "no client initialized for api key"
	}

	return ""
}

