package wasm

import (
	"math"

	"cosmossdk.io/errors"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	types "github.com/CosmWasm/wasmd/x/wasm/types"
)

var _ porttypes.IBCModule = IBCHandler{}

type IBCHandler struct {
	keeper        types.IBCContractKeeper
	channelKeeper types.ChannelKeeper
}

func NewIBCHandler(k types.IBCContractKeeper, ck types.ChannelKeeper) IBCHandler {
	return IBCHandler{keeper: k, channelKeeper: ck}
}

// OnChanOpenInit implements the IBCModule interface
func (i IBCHandler) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterParty channeltypes.Counterparty,
	version string,
) (string, error) {
	// ensure port, version, capability
	if err := ValidateChannelParams(channelID); err != nil {
		return "", err
	}
	contractAddr, err := ContractFromPortID(portID)
	if err != nil {
		return "", errors.Wrapf(err, "contract port id")
	}

	msg := wasmvmtypes.IBCChannelOpenMsg{
		OpenInit: &wasmvmtypes.IBCOpenInit{
			Channel: wasmvmtypes.IBCChannel{
				Endpoint:             wasmvmtypes.IBCEndpoint{PortID: portID, ChannelID: channelID},
				CounterpartyEndpoint: wasmvmtypes.IBCEndpoint{PortID: counterParty.PortId, ChannelID: counterParty.ChannelId},
				Order:                order.String(),
				Version:              version,
				ConnectionID:         connectionHops[0], // At the moment this list must be of length 1. In the future multi-hop channels may be supported.
			},
		},
	}
	err = i.keeper.OnOpenChannel(ctx, contractAddr, msg)
	if err != nil {
		return "", err
	}
	return version, nil
}

// OnChanOpenTry implements the IBCModule interface
func (i IBCHandler) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID, channelID string,
	chanCap *capabilitytypes.Capability,
	counterParty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	// ensure port, version, capability
	if err := ValidateChannelParams(channelID); err != nil {
		return "", err
	}

	contractAddr, err := ContractFromPortID(portID)
	if err != nil {
		return "", errors.Wrapf(err, "contract port id")
	}

	msg := wasmvmtypes.IBCChannelOpenMsg{
		OpenTry: &wasmvmtypes.IBCOpenTry{
			Channel: wasmvmtypes.IBCChannel{
				Endpoint:             wasmvmtypes.IBCEndpoint{PortID: portID, ChannelID: channelID},
				CounterpartyEndpoint: wasmvmtypes.IBCEndpoint{PortID: counterParty.PortId, ChannelID: counterParty.ChannelId},
				Order:                order.String(),
				Version:              counterpartyVersion,
				ConnectionID:         connectionHops[0], // At the moment this list must be of length 1. In the future multi-hop channels may be supported.
			},
			CounterpartyVersion: counterpartyVersion,
		},
	}

	err = i.keeper.OnOpenChannel(ctx, contractAddr, msg)
	if err != nil {
		return "", err
	}
	return counterpartyVersion, nil
}

// OnChanOpenAck implements the IBCModule interface
func (i IBCHandler) OnChanOpenAck(
	ctx sdk.Context,
	portID, channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	contractAddr, err := ContractFromPortID(portID)
	if err != nil {
		return errors.Wrapf(err, "contract port id")
	}
	channelInfo, ok := i.channelKeeper.GetChannel(ctx, portID, channelID)
	if !ok {
		return errors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}
	msg := wasmvmtypes.IBCChannelConnectMsg{
		OpenAck: &wasmvmtypes.IBCOpenAck{
			Channel:             toWasmVMChannel(portID, channelID, channelInfo),
			CounterpartyVersion: counterpartyVersion,
		},
	}
	return i.keeper.OnConnectChannel(ctx, contractAddr, msg)
}

// OnChanOpenConfirm implements the IBCModule interface
func (i IBCHandler) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	contractAddr, err := ContractFromPortID(portID)
	if err != nil {
		return errors.Wrapf(err, "contract port id")
	}
	channelInfo, ok := i.channelKeeper.GetChannel(ctx, portID, channelID)
	if !ok {
		return errors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}
	msg := wasmvmtypes.IBCChannelConnectMsg{
		OpenConfirm: &wasmvmtypes.IBCOpenConfirm{
			Channel: toWasmVMChannel(portID, channelID, channelInfo),
		},
	}
	return i.keeper.OnConnectChannel(ctx, contractAddr, msg)
}

// OnChanCloseInit implements the IBCModule interface
func (i IBCHandler) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	contractAddr, err := ContractFromPortID(portID)
	if err != nil {
		return errors.Wrapf(err, "contract port id")
	}
	channelInfo, ok := i.channelKeeper.GetChannel(ctx, portID, channelID)
	if !ok {
		return errors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	msg := wasmvmtypes.IBCChannelCloseMsg{
		CloseInit: &wasmvmtypes.IBCCloseInit{Channel: toWasmVMChannel(portID, channelID, channelInfo)},
	}
	err = i.keeper.OnCloseChannel(ctx, contractAddr, msg)
	if err != nil {
		return err
	}
	// emit events?

	return err
}

// OnChanCloseConfirm implements the IBCModule interface
func (i IBCHandler) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	// counterparty has closed the channel
	contractAddr, err := ContractFromPortID(portID)
	if err != nil {
		return errors.Wrapf(err, "contract port id")
	}
	channelInfo, ok := i.channelKeeper.GetChannel(ctx, portID, channelID)
	if !ok {
		return errors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	msg := wasmvmtypes.IBCChannelCloseMsg{
		CloseConfirm: &wasmvmtypes.IBCCloseConfirm{Channel: toWasmVMChannel(portID, channelID, channelInfo)},
	}
	err = i.keeper.OnCloseChannel(ctx, contractAddr, msg)
	if err != nil {
		return err
	}
	// emit events?

	return err
}

func toWasmVMChannel(portID, channelID string, channelInfo channeltypes.Channel) wasmvmtypes.IBCChannel {
	return wasmvmtypes.IBCChannel{
		Endpoint:             wasmvmtypes.IBCEndpoint{PortID: portID, ChannelID: channelID},
		CounterpartyEndpoint: wasmvmtypes.IBCEndpoint{PortID: channelInfo.Counterparty.PortId, ChannelID: channelInfo.Counterparty.ChannelId},
		Order:                channelInfo.Ordering.String(),
		Version:              channelInfo.Version,
		ConnectionID:         channelInfo.ConnectionHops[0], // At the moment this list must be of length 1. In the future multi-hop channels may be supported.
	}
}

// OnRecvPacket implements the IBCModule interface
func (i IBCHandler) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	contractAddr, err := ContractFromPortID(packet.DestinationPort)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(errors.Wrapf(err, "contract port id").Error())
	}
	msg := wasmvmtypes.IBCPacketReceiveMsg{Packet: newIBCPacket(packet)}
	ack, err := i.keeper.OnRecvPacket(ctx, contractAddr, msg)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}
	return ContractConfirmStateAck(ack)
}

var _ ibcexported.Acknowledgement = ContractConfirmStateAck{}

type ContractConfirmStateAck []byte

func (w ContractConfirmStateAck) Success() bool {
	return true // always commit state
}

func (w ContractConfirmStateAck) Acknowledgement() []byte {
	return w
}

// OnAcknowledgementPacket implements the IBCModule interface
func (i IBCHandler) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	contractAddr, err := ContractFromPortID(packet.SourcePort)
	if err != nil {
		return errors.Wrapf(err, "contract port id")
	}

	err = i.keeper.OnAckPacket(ctx, contractAddr, wasmvmtypes.IBCPacketAckMsg{
		Acknowledgement: wasmvmtypes.IBCAcknowledgement{Data: acknowledgement},
		OriginalPacket:  newIBCPacket(packet),
	})
	if err != nil {
		return errors.Wrap(err, "on ack")
	}
	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (i IBCHandler) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	contractAddr, err := ContractFromPortID(packet.SourcePort)
	if err != nil {
		return errors.Wrapf(err, "contract port id")
	}
	msg := wasmvmtypes.IBCPacketTimeoutMsg{Packet: newIBCPacket(packet)}
	err = i.keeper.OnTimeoutPacket(ctx, contractAddr, msg)
	if err != nil {
		return errors.Wrap(err, "on timeout")
	}
	return nil
}

func (i IBCHandler) NegotiateAppVersion(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionID string,
	portID string,
	counterparty channeltypes.Counterparty,
	proposedVersion string,
) (version string, err error) {
	return proposedVersion, nil // accept all
}

func newIBCPacket(packet channeltypes.Packet) wasmvmtypes.IBCPacket {
	timeout := wasmvmtypes.IBCTimeout{
		Timestamp: packet.TimeoutTimestamp,
	}
	if !packet.TimeoutHeight.IsZero() {
		timeout.Block = &wasmvmtypes.IBCTimeoutBlock{
			Height:   packet.TimeoutHeight.RevisionHeight,
			Revision: packet.TimeoutHeight.RevisionNumber,
		}
	}

	return wasmvmtypes.IBCPacket{
		Data:     packet.Data,
		Src:      wasmvmtypes.IBCEndpoint{ChannelID: packet.SourceChannel, PortID: packet.SourcePort},
		Dest:     wasmvmtypes.IBCEndpoint{ChannelID: packet.DestinationChannel, PortID: packet.DestinationPort},
		Sequence: packet.Sequence,
		Timeout:  timeout,
	}
}

func ValidateChannelParams(channelID string) error {
	// NOTE: for escrow address security only 2^32 channels are allowed to be created
	// Issue: https://github.com/cosmos/cosmos-sdk/issues/7737
	channelSequence, err := channeltypes.ParseChannelSequence(channelID)
	if err != nil {
		return err
	}
	if channelSequence > math.MaxUint32 {
		return errors.Wrapf(types.ErrMaxIBCChannels, "channel sequence %d is greater than max allowed transfer channels %d", channelSequence, math.MaxUint32)
	}
	return nil
}
