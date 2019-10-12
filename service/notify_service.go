package service

import (
	"gitlab/nefco/auction/core"
	"gitlab/nefco/auction/server/command"
)

const (
	eventLotAdded   = "event.lot.added"
	eventLotChanged = "event.lot.changed"
	eventLotDeleted = "event.lot.deleted"
)

type Notify struct {
	Event    command.Event
	Receiver *core.User
	lot      *core.Lot
}

func (n *Notify) Update(receiverID uint) {
	if n.lot == nil {
		return
	}
	n.lot.UpdatePrice(receiverID)
}

func (n *Notify) Check(receiver *core.User) bool {
	if receiver.Blocked {
		return false
	}
	if n.lot == nil {
		return true
	}
	return receiver.CheckFilter(n.lot)
}

type notifyService struct {
	ch chan Notify
}

func NewNotifyService() *notifyService {
	return &notifyService{
		ch: make(chan Notify),
	}
}

func (svc *notifyService) add(evt command.Event, lot *core.Lot) {
	svc.ch <- Notify{Event: evt, lot: lot}
}

func (svc *notifyService) addWithReceiver(evt command.Event, receiver *core.User) {
	svc.ch <- Notify{Event: evt, Receiver: receiver}
}

func (svc *notifyService) Events() <-chan Notify {
	return svc.ch
}

func (svc *notifyService) LotAdded(lot *core.Lot) {
	svc.add(command.LotEvent(eventLotAdded, lot), lot)
}

func (svc *notifyService) LotChanged(lot core.Lot) {
	l := &lot
	svc.add(command.LotEvent(eventLotChanged, l), l)
}

func (svc *notifyService) LotChangedWithReceiver(lot *core.Lot, receiver *core.User) {
	svc.addWithReceiver(command.LotEvent(eventLotChanged, lot), receiver)
}

func (svc *notifyService) LotDeleted(lot *core.Lot) {
	svc.add(command.LotEvent(eventLotDeleted, lot), lot)
}
