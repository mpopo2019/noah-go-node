package api

import (
	"github.com/noah-blockchain/noah-go-node/core/commissions"
	"github.com/noah-blockchain/noah-go-node/core/transaction"
	"github.com/noah-blockchain/noah-go-node/core/types"
	"github.com/noah-blockchain/noah-go-node/formula"
	"github.com/noah-blockchain/noah-go-node/rpc/lib/types"
	"math/big"
)

type EstimateCoinSellAllResponse struct {
	WillGet string `json:"will_get"`
}

func EstimateCoinSellAll(
	coinToSellString string, coinToBuyString string, valueToSell *big.Int, gasPrice uint64, height int) (*EstimateCoinSellAllResponse,
	error) {
	cState, err := GetStateForHeight(height)
	if err != nil {
		return nil, err
	}

	if gasPrice < 1 {
		gasPrice = 1
	}

	coinToSell := types.StrToCoinSymbol(coinToSellString)
	coinToBuy := types.StrToCoinSymbol(coinToBuyString)

	var result *big.Int

	if coinToSell == coinToBuy {
		return nil, rpctypes.RPCError{Code: 400, Message: "\"From\" coin equals to \"to\" coin"}
	}

	if !cState.CoinExists(coinToSell) {
		return nil, rpctypes.RPCError{Code: 404, Message: "Coin to sell not exists"}
	}

	if !cState.CoinExists(coinToBuy) {
		return nil, rpctypes.RPCError{Code: 404, Message: "Coin to buy not exists"}
	}

	commissionInBaseCoin := big.NewInt(commissions.ConvertTx)
	commissionInBaseCoin.Mul(commissionInBaseCoin, transaction.CommissionMultiplier)
	commission := big.NewInt(0).Set(commissionInBaseCoin)

	switch {
	case coinToSell == types.GetBaseCoin():
		coin := cState.GetStateCoin(coinToBuy).Data()

		valueToSell.Sub(valueToSell, commission)
		if valueToSell.Cmp(big.NewInt(0)) != 1 {
			return nil, rpctypes.RPCError{Code: 400, Message: "Not enough coins to pay commission"}
		}

		result = formula.CalculatePurchaseReturn(coin.Volume, coin.ReserveBalance, coin.Crr, valueToSell)
	case coinToBuy == types.GetBaseCoin():
		coin := cState.GetStateCoin(coinToSell).Data()
		result = formula.CalculateSaleReturn(coin.Volume, coin.ReserveBalance, coin.Crr, valueToSell)

		result.Sub(result, commission)
		if result.Cmp(big.NewInt(0)) != 1 {
			return nil, rpctypes.RPCError{Code: 400, Message: "Not enough coins to pay commission"}
		}
	default:
		coinFrom := cState.GetStateCoin(coinToSell).Data()
		coinTo := cState.GetStateCoin(coinToBuy).Data()
		basecoinValue := formula.CalculateSaleReturn(coinFrom.Volume, coinFrom.ReserveBalance, coinFrom.Crr, valueToSell)

		basecoinValue.Sub(basecoinValue, commission)
		if basecoinValue.Cmp(big.NewInt(0)) != 1 {
			return nil, rpctypes.RPCError{Code: 400, Message: "Not enough coins to pay commission"}
		}

		result = formula.CalculatePurchaseReturn(coinTo.Volume, coinTo.ReserveBalance, coinTo.Crr, basecoinValue)
	}

	return &EstimateCoinSellAllResponse{
		WillGet: result.String(),
	}, nil
}
