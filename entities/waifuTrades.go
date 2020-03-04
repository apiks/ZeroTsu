package entities

import "sync"

type WaifuTrade struct {
	sync.RWMutex

	TradeID     string `json:"TradeID"`
	InitiatorID string `json:"InitiatorID"`
	AccepteeID  string `json:"AccepteeID"`
}

func NewWaifuTrade(tradeID string, initiatorID string, accepteeID string) *WaifuTrade {
	return &WaifuTrade{TradeID: tradeID, InitiatorID: initiatorID, AccepteeID: accepteeID}
}

func (w *WaifuTrade) SetTradeID(tradeID string) {
	w.Lock()
	w.TradeID = tradeID
	w.Unlock()
}

func (w *WaifuTrade) GetTradeID() string {
	w.RLock()
	defer w.RUnlock()
	if w == nil {
		return ""
	}
	return w.TradeID
}

func (w *WaifuTrade) SetInitiatorID(initiatorID string) {
	w.Lock()
	w.InitiatorID = initiatorID
	w.Unlock()
}

func (w *WaifuTrade) GetInitiatorID() string {
	w.RLock()
	defer w.RUnlock()
	if w == nil {
		return ""
	}
	return w.InitiatorID
}

func (w *WaifuTrade) SetAccepteeID(accepteeID string) {
	w.Lock()
	w.AccepteeID = accepteeID
	w.Unlock()
}

func (w *WaifuTrade) GetAccepteeID() string {
	w.RLock()
	defer w.RUnlock()
	if w == nil {
		return ""
	}
	return w.AccepteeID
}
