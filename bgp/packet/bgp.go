// bgp.go
package packet

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
)

type BGPPktInfo struct {
	Msg      *BGPMessage
	MsgError *BGPMessageError
}

func NewBGPPktInfo(msg *BGPMessage, msgError *BGPMessageError) *BGPPktInfo {
	return &BGPPktInfo{msg, msgError}
}

type BGPPktSrc struct {
	Src string
	Msg *BGPMessage
}

func NewBGPPktSrc(src string, msg *BGPMessage) *BGPPktSrc {
	return &BGPPktSrc{src, msg}
}

type BGPPeerAttrs struct {
	ASSize uint8
}

const BGPASTrans uint16 = 23456

const BGPHeaderMarkerLen int = 16

const (
	_ uint8 = iota
	BGPMsgTypeOpen
	BGPMsgTypeUpdate
	BGPMsgTypeNotification
	BGPMsgTypeKeepAlive
)

const (
	BGPMsgHeaderLen = 19
	BGPMsgMaxLen    = 4096
)

const (
	_ uint8 = iota
	BGPMsgHeaderError
	BGPOpenMsgError
	BGPUpdateMsgError
	BGPHoldTimerExpired
	BGPFSMError
	BGPCease
)

const (
	_ uint8 = iota
	BGPConnNotSychd
	BGPBadMessageLen
	BGPBadMessageType
)

const (
	BGPUnspecific uint8 = iota
	BGPUnsupportedVersionNumber
	BGPBadPeerAS
	BGPBadBGPIdentifier
	BGPUnsupportedOptionalParam
	_
	BGPUnacceptableHoldTime
	BGPUnsupportedCapability
)

const (
	_ uint8 = iota
	BGPMalformedAttrList
	BGPUnrecognizedWellKnownAttr
	BGPMissingWellKnownAttr
	BGPAttrFlagsError
	BGPAttrLenError
	BGPInvalidOriginAttr
	_
	BGPInvalidNextHopAttr
	BGPOptionalAttrError
	BGPInvalidNetworkField
	BGPMalformedASPath
)

type BGPOptParamType uint8

const (
	_ BGPOptParamType = iota
	_
	BGPOptParamTypeCapability
)

var BGPOptParamTypeToStruct = map[BGPOptParamType]BGPOptParam{
	BGPOptParamTypeCapability: &BGPOptParamCapability{},
}

type BGPCapabilityType uint8

const (
	_                 BGPCapabilityType = iota
	BGPCapTypeAS4Path                   = 65
)

var BGPCapTypeToStruct = map[BGPCapabilityType]BGPCapability{
	BGPCapTypeAS4Path: &BGPCapAS4Path{},
}

type BGPPathAttrFlag uint8

const (
	_ BGPPathAttrFlag = 1 << (iota + 3)
	BGPPathAttrFlagExtendedLen
	BGPPathAttrFlagPartial
	BGPPathAttrFlagTransitive
	BGPPathAttrFlagOptional
)

var BGPPathAttrFlagAll BGPPathAttrFlag = 0xF0
var BGPPathAttrFlagAllMinusExtendedLen BGPPathAttrFlag = 0xE0

type BGPPathAttrType uint8

const (
	_ BGPPathAttrType = iota
	BGPPathAttrTypeOrigin
	BGPPathAttrTypeASPath
	BGPPathAttrTypeNextHop
	BGPPathAttrTypeMultiExitDisc
	BGPPathAttrTypeLocalPref
	BGPPathAttrTypeAtomicAggregate
	BGPPathAttrTypeAggregator
	_
	BGPPathAttrTypeOriginatorId
	BGPPathAttrTypeClusterList
	BGPPathAttrTypeUnknown
	_
	_
	_
	_
	_
	BGPPathAttrTypeAS4Path
	BGPPathAttrTypeAS4Aggregator
)

type BGPPathAttrOriginType uint8

const (
	BGPPathAttrOriginIGP BGPPathAttrOriginType = iota
	BGPPathAttrOriginEGP
	BGPPathAttrOriginIncomplete
	BGPPathAttrOriginMax
)

type BGPASPathSegmentType uint8

const (
	BGPASPathSet BGPASPathSegmentType = iota + 1
	BGPASPathSequence
)

var BGPPathAttrWellKnownMandatory = []BGPPathAttrType{
	BGPPathAttrTypeOrigin, BGPPathAttrTypeASPath, BGPPathAttrTypeNextHop}

var BGPPathAttrTypeToStructMap = map[BGPPathAttrType]BGPPathAttr{
	BGPPathAttrTypeOrigin:          &BGPPathAttrOrigin{},
	BGPPathAttrTypeASPath:          &BGPPathAttrASPath{},
	BGPPathAttrTypeNextHop:         &BGPPathAttrNextHop{},
	BGPPathAttrTypeMultiExitDisc:   &BGPPathAttrMultiExitDisc{},
	BGPPathAttrTypeLocalPref:       &BGPPathAttrLocalPref{},
	BGPPathAttrTypeAtomicAggregate: &BGPPathAttrAtomicAggregate{},
	BGPPathAttrTypeAggregator:      &BGPPathAttrAggregator{},
}

var BGPPathAttrTypeFlagsMap = map[BGPPathAttrType][]BGPPathAttrFlag{
	BGPPathAttrTypeOrigin:          []BGPPathAttrFlag{BGPPathAttrFlagTransitive, BGPPathAttrFlagAllMinusExtendedLen},
	BGPPathAttrTypeASPath:          []BGPPathAttrFlag{BGPPathAttrFlagTransitive, BGPPathAttrFlagAllMinusExtendedLen},
	BGPPathAttrTypeNextHop:         []BGPPathAttrFlag{BGPPathAttrFlagTransitive, BGPPathAttrFlagAllMinusExtendedLen},
	BGPPathAttrTypeMultiExitDisc:   []BGPPathAttrFlag{BGPPathAttrFlagOptional, BGPPathAttrFlagAllMinusExtendedLen},
	BGPPathAttrTypeLocalPref:       []BGPPathAttrFlag{BGPPathAttrFlagTransitive, BGPPathAttrFlagAllMinusExtendedLen},
	BGPPathAttrTypeAtomicAggregate: []BGPPathAttrFlag{BGPPathAttrFlagTransitive, BGPPathAttrFlagAllMinusExtendedLen},
	BGPPathAttrTypeAggregator:      []BGPPathAttrFlag{BGPPathAttrFlagOptional & BGPPathAttrFlagTransitive, BGPPathAttrFlagAllMinusExtendedLen},
}

var BGPPathAttrTypeLenMap = map[BGPPathAttrType]uint16{
	BGPPathAttrTypeOrigin:          1,
	BGPPathAttrTypeNextHop:         4,
	BGPPathAttrTypeMultiExitDisc:   4,
	BGPPathAttrTypeLocalPref:       4,
	BGPPathAttrTypeAtomicAggregate: 0,
	BGPPathAttrTypeAggregator:      6,
}

type BGPMessageError struct {
	TypeCode    uint8
	SubTypeCode uint8
	Data        []byte
	Message     string
}

func (e BGPMessageError) Error() string {
	return fmt.Sprintf("%v:%v - %v", e.TypeCode, e.SubTypeCode, e.Message)
}

type BGPHeader struct {
	Marker [BGPHeaderMarkerLen]byte
	Length uint16
	Type   uint8
}

func (header *BGPHeader) Clone() *BGPHeader {
	x := *header
	return &x
}

func (header *BGPHeader) Encode() ([]byte, error) {
	pkt := make([]byte, 19)
	for i := 0; i < BGPHeaderMarkerLen; i++ {
		pkt[i] = 0xff
	}
	binary.BigEndian.PutUint16(pkt[16:18], header.Length)
	pkt[18] = header.Type
	return pkt, nil
}

func (header *BGPHeader) Decode(pkt []byte) error {
	header.Length = binary.BigEndian.Uint16(pkt[16:18])
	header.Type = pkt[18]
	return nil
}

func (header *BGPHeader) Len() uint32 {
	return uint32(header.Length)
}

type BGPBody interface {
	Clone() BGPBody
	Encode() ([]byte, error)
	Decode(*BGPHeader, []byte, interface{}) error
}

type BGPCapability interface {
	Encode() ([]byte, error)
	Decode(pkt []byte) error
	TotalLen() uint8
	GetCode() BGPCapabilityType
}

type BGPCapabilityBase struct {
	Type BGPCapabilityType
	Len  uint8
}

func (msg *BGPCapabilityBase) Encode() ([]byte, error) {
	pkt := make([]byte, msg.TotalLen())
	pkt[0] = uint8(msg.Type)
	pkt[1] = msg.Len
	return pkt, nil
}

func (msg *BGPCapabilityBase) Decode(pkt []byte) error {
	if len(pkt) < 2 {
		return BGPMessageError{BGPOpenMsgError, BGPUnspecific, nil, "Not enough data to decode capability type and length"}
	}

	msg.Type = BGPCapabilityType(pkt[0])
	msg.Len = pkt[1]

	if len(pkt) < int(msg.TotalLen()) {
		return BGPMessageError{BGPUpdateMsgError, BGPUnspecific, nil, "Not enough data to decode capability data"}
	}
	return nil
}

func (msg *BGPCapabilityBase) TotalLen() uint8 {
	return msg.Len + 2
}

func (msg *BGPCapabilityBase) GetCode() BGPCapabilityType {
	return msg.Type
}

type BGPCapAS4Path struct {
	BGPCapabilityBase
	Value uint32
}

func (msg *BGPCapAS4Path) Encode() ([]byte, error) {
	pkt, err := msg.BGPCapabilityBase.Encode()
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint32(pkt[2:], msg.Value)
	return pkt, nil
}

func (msg *BGPCapAS4Path) Decode(pkt []byte) error {
	err := msg.BGPCapabilityBase.Decode(pkt)
	if err != nil {
		return err
	}

	msg.Value = binary.BigEndian.Uint32(pkt[2:])
	return nil
}

func NewBGPCap4ByteASPath(as uint32) *BGPCapAS4Path {
	return &BGPCapAS4Path{
		BGPCapabilityBase: BGPCapabilityBase{
			Type: BGPCapTypeAS4Path,
			Len:  4,
		},
		Value: as,
	}
}

type BGPCapUnknown struct {
	BGPCapabilityBase
	Value []byte
}

func (msg *BGPCapUnknown) Encode() ([]byte, error) {
	pkt, err := msg.BGPCapabilityBase.Encode()
	if err != nil {
		return nil, err
	}

	copy(pkt[2:], msg.Value)
	return pkt, nil
}

func (msg *BGPCapUnknown) Decode(pkt []byte) error {
	err := msg.BGPCapabilityBase.Decode(pkt)
	if err != nil {
		return err
	}

	copy(msg.Value, pkt[2:])
	return nil
}

type BGPOptParam interface {
	Encode() ([]byte, error)
	Decode(pkt []byte) error
	TotalLen() uint8
	GetCode() BGPOptParamType
}

type BGPOptParamBase struct {
	Type BGPOptParamType
	Len  uint8
}

func (msg *BGPOptParamBase) Encode() ([]byte, error) {
	pkt := make([]byte, 2)
	pkt[0] = uint8(msg.Type)
	pkt[1] = msg.Len
	return pkt, nil
}

func (msg *BGPOptParamBase) Decode(pkt []byte) error {
	if len(pkt) < 2 {
		return BGPMessageError{BGPOpenMsgError, BGPUnspecific, nil, "Not enough data to decode Opt params type and length"}
	}

	msg.Type = BGPOptParamType(pkt[0])
	msg.Len = pkt[1]

	if len(pkt) < int(msg.TotalLen()) {
		return BGPMessageError{BGPUpdateMsgError, BGPUnspecific, nil, "Not enough data to decode Opt params data"}
	}
	return nil
}

func (msg *BGPOptParamBase) TotalLen() uint8 {
	return msg.Len + 2
}

func (msg *BGPOptParamBase) GetCode() BGPOptParamType {
	return msg.Type
}

type BGPOptParamCapability struct {
	BGPOptParamBase
	Value []BGPCapability
}

func (msg *BGPOptParamCapability) Encode() ([]byte, error) {
	pkt, err := msg.BGPOptParamBase.Encode()
	if err != nil {
		return nil, err
	}

	for _, capability := range msg.Value {
		bytes, err := capability.Encode()
		if err != nil {
			return nil, err
		}

		pkt = append(pkt, bytes...)
	}
	return pkt, nil
}

func (msg *BGPOptParamCapability) GetCapParam(pkt []byte) BGPCapability {
	capParamType := BGPCapabilityType(pkt[0])
	if capParam, ok := BGPCapTypeToStruct[capParamType]; ok {
		return capParam
	} else {
		return &BGPCapUnknown{}
	}
}

func (msg *BGPOptParamCapability) Decode(pkt []byte) error {
	err := msg.BGPOptParamBase.Decode(pkt)
	if err != nil {
		return err
	}

	paramsLen := msg.Len
	msg.Value = make([]BGPCapability, 0)
	offset := uint8(2)
	for paramsLen > 0 {
		capParam := msg.GetCapParam(pkt[offset:])
		if err != nil {
			return err
		}

		err = capParam.Decode(pkt[offset:])
		if err != nil {
			return err
		}
		msg.Value = append(msg.Value, capParam)
		offset += capParam.TotalLen()
		paramsLen -= capParam.TotalLen()
	}
	return nil
}

func NewBGPOptParamCapability(capabilities []BGPCapability) *BGPOptParamCapability {
	paramsLen := uint8(0)
	for _, capability := range capabilities {
		paramsLen = capability.TotalLen()
	}

	return &BGPOptParamCapability{
		BGPOptParamBase: BGPOptParamBase{
			Type: BGPOptParamTypeCapability,
			Len:  paramsLen,
		},
		Value: capabilities,
	}
}

type BGPOptParamUnknown struct {
	BGPOptParamBase
	Value []byte
}

func (msg *BGPOptParamUnknown) Encode() ([]byte, error) {
	pkt, err := msg.BGPOptParamBase.Encode()
	if err != nil {
		return nil, err
	}

	copy(pkt[2:], msg.Value)
	return pkt, nil
}

func (msg *BGPOptParamUnknown) Decode(pkt []byte) error {
	err := msg.BGPOptParamBase.Decode(pkt)
	if err != nil {
		return err
	}

	msg.Value = make([]byte, msg.Len)
	copy(msg.Value, pkt[2:])
	return nil
}

type BGPOpen struct {
	Version     uint8
	MyAS        uint32
	HoldTime    uint16
	BGPId       net.IP
	OptParamLen uint8
	OptParams   []BGPOptParam
}

func (msg *BGPOpen) Clone() BGPBody {
	x := *msg
	x.BGPId = make(net.IP, len(x.BGPId), cap(x.BGPId))
	copy(x.BGPId, msg.BGPId)
	return &x
}

func (msg *BGPOpen) Encode() ([]byte, error) {
	pkt := make([]byte, 10)
	pkt[0] = msg.Version
	binary.BigEndian.PutUint16(pkt[1:3], uint16(msg.MyAS))
	binary.BigEndian.PutUint16(pkt[3:5], msg.HoldTime)
	copy(pkt[5:9], msg.BGPId.To4())
	paramsLen := uint8(0)
	for _, param := range msg.OptParams {
		bytes, err := param.Encode()
		if err != nil {
			return nil, err
		}

		pkt = append(pkt, bytes...)
		paramsLen += param.TotalLen()
	}
	pkt[9] = paramsLen
	return pkt, nil
}

func (msg *BGPOpen) GetOptParam(pkt []byte) (BGPOptParam, error) {
	optParamType := BGPOptParamType(pkt[0])
	if optParam, ok := BGPOptParamTypeToStruct[optParamType]; ok {
		return optParam, nil
	} else {
		return &BGPOptParamUnknown{}, BGPMessageError{BGPOpenMsgError, BGPUnsupportedOptionalParam, nil,
			fmt.Sprintf("Unknown optional parameter %d", optParamType)}
	}
}

func (msg *BGPOpen) Decode(header *BGPHeader, pkt []byte, data interface{}) error {
	msg.Version = pkt[0]
	msg.MyAS = uint32(binary.BigEndian.Uint16(pkt[1:3]))
	msg.HoldTime = binary.BigEndian.Uint16(pkt[3:5])
	msg.BGPId = net.IP(pkt[5:9]).To4()
	msg.OptParamLen = pkt[9]

	msg.OptParams = make([]BGPOptParam, 0)
	paramsLen := msg.OptParamLen
	offset := uint8(10)
	for paramsLen > 0 {
		optParam, err := msg.GetOptParam(pkt[offset:])
		if err != nil {
			return err
		}
		err = optParam.Decode(pkt[offset:])
		if err != nil {
			return err
		}
		msg.OptParams = append(msg.OptParams, optParam)
		offset += optParam.TotalLen()
		paramsLen -= optParam.TotalLen()
	}
	return nil
}

func NewBGPOpenMessage(myAS uint32, holdTime uint16, bgpId string, optParams []BGPOptParam) *BGPMessage {
	optParamsLen := uint8(0)
	for _, param := range optParams {
		optParamsLen += param.TotalLen()
	}

	if myAS > math.MaxUint16 {
		myAS = uint32(BGPASTrans)
	}

	return &BGPMessage{
		Header: BGPHeader{Type: BGPMsgTypeOpen},
		Body:   &BGPOpen{4, myAS, holdTime, net.ParseIP(bgpId), optParamsLen, optParams},
	}
}

type BGPKeepAlive struct {
}

func (msg *BGPKeepAlive) Clone() BGPBody {
	x := *msg
	return &x
}

func (msg *BGPKeepAlive) Encode() ([]byte, error) {
	return nil, nil
}

func (msg *BGPKeepAlive) Decode(*BGPHeader, []byte, interface{}) error {
	return nil
}

func NewBGPKeepAliveMessage() *BGPMessage {
	return &BGPMessage{
		Header: BGPHeader{Length: 19, Type: BGPMsgTypeKeepAlive},
		Body:   &BGPKeepAlive{},
	}
}

type BGPNotification struct {
	ErrorCode    uint8
	ErrorSubcode uint8
	Data         []byte
}

func (msg *BGPNotification) Clone() BGPBody {
	x := *msg
	x.Data = make([]byte, len(msg.Data), cap(msg.Data))
	copy(x.Data, msg.Data)
	return &x
}

func (msg *BGPNotification) Encode() ([]byte, error) {
	pkt := make([]byte, 2)
	pkt[0] = msg.ErrorCode
	pkt[1] = msg.ErrorSubcode
	pkt = append(pkt, msg.Data...)
	return pkt, nil
}

func (msg *BGPNotification) Decode(header *BGPHeader, pkt []byte, data interface{}) error {
	msg.ErrorCode = pkt[0]
	msg.ErrorSubcode = pkt[1]
	if len(pkt) > 2 {
		msg.Data = pkt[2:]
	}
	return nil
}

func NewBGPNotificationMessage(errorCode uint8, errorSubCode uint8, data []byte) *BGPMessage {
	return &BGPMessage{
		Header: BGPHeader{Length: 21 + uint16(len(data)), Type: BGPMsgTypeNotification},
		Body:   &BGPNotification{errorCode, errorSubCode, data},
	}
}

type IPPrefix struct {
	Length uint8
	Prefix net.IP
}

func (ip *IPPrefix) Clone() *IPPrefix {
	x := *ip
	x.Prefix = make(net.IP, len(ip.Prefix), cap(ip.Prefix))
	copy(x.Prefix, ip.Prefix)
	return &x
}

func (ip *IPPrefix) Encode() ([]byte, error) {
	pkt := make([]byte, ip.Len())
	pkt[0] = ip.Length
	ipBytesStart := uint8(cap(ip.Prefix) - 4)
	copy(pkt[1:], ip.Prefix[ipBytesStart:ipBytesStart+((ip.Length+7)/8)])
	return pkt, nil
}

func (ip *IPPrefix) Decode(pkt []byte) error {
	ip.Length = pkt[0]
	bytes := (ip.Length + 7) / 8
	if len(pkt) < int(bytes) {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Prefix length invalid"}
	}
	ip.Prefix = make(net.IP, 4)
	copy(ip.Prefix, pkt[1:bytes+1])
	return nil
}

func (ip *IPPrefix) Len() uint32 {
	return uint32(((ip.Length + 7) / 8) + 1)
}

func NewIPPrefix(prefix net.IP, length uint8) *IPPrefix {
	return &IPPrefix{
		Length: length,
		Prefix: prefix,
	}
}

type BGPPathAttr interface {
	Clone() BGPPathAttr
	Encode() ([]byte, error)
	Decode(pkt []byte, data interface{}) error
	TotalLen() uint32
	GetCode() BGPPathAttrType
}

type BGPPathAttrBase struct {
	Flags          BGPPathAttrFlag
	Code           BGPPathAttrType
	Length         uint16
	BGPPathAttrLen uint16
}

func (pa *BGPPathAttrBase) Clone() BGPPathAttrBase {
	x := *pa
	return x
}

func (pa *BGPPathAttrBase) Encode() ([]byte, error) {
	pkt := make([]byte, pa.TotalLen())
	pkt[0] = uint8(pa.Flags)
	pkt[1] = uint8(pa.Code)

	if pa.Flags&BGPPathAttrFlagExtendedLen != 0 {
		binary.BigEndian.PutUint16(pkt[2:], pa.Length)
	} else {
		pkt[2] = uint8(pa.Length)
	}

	return pkt, nil
}

func (pa *BGPPathAttrBase) checkFlags(pkt []byte) error {
	if pa.Flags&BGPPathAttrFlagOptional != 0 &&
		pa.Flags&BGPPathAttrFlagTransitive == 0 &&
		pa.Flags&BGPPathAttrFlagPartial == 0 {
		return BGPMessageError{BGPUpdateMsgError, BGPAttrFlagsError, pkt[:pa.TotalLen()],
			"Partial bit in a optional transitive attr is not set"}
	}

	return nil
}

func (pa *BGPPathAttrBase) Decode(pkt []byte, data interface{}) error {
	if len(pkt) < 3 {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Not enough data to decode"}
	}

	pa.Flags = BGPPathAttrFlag(pkt[0])
	pa.Code = BGPPathAttrType(pkt[1])

	if pa.Flags&BGPPathAttrFlagExtendedLen != 0 {
		if len(pkt) < 4 {
			return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Not enough data to decode"}
		}
		pa.Length = binary.BigEndian.Uint16(pkt[2:4])
		pa.BGPPathAttrLen = 4
	} else {
		pa.Length = uint16(pkt[2])
		pa.BGPPathAttrLen = 3
	}
	if len(pkt) < int(pa.Length+pa.BGPPathAttrLen) {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, pkt, "Not enough data to decode"}
	}

	val, ok := BGPPathAttrTypeFlagsMap[pa.Code]
	if ok {
		if (val[0]^pa.Flags)&val[1] != 0 {
			return BGPMessageError{BGPUpdateMsgError, BGPAttrFlagsError, pkt[:pa.TotalLen()], "Bad Attribute Flags"}
		}
	}

	err := pa.checkFlags(pkt)
	if err != nil {
		return err
	}

	length, ok := BGPPathAttrTypeLenMap[pa.Code]
	if ok {
		if length != pa.Length {
			return BGPMessageError{BGPUpdateMsgError, BGPAttrLenError, pkt[:pa.TotalLen()], "Bad Attribute Length"}
		}
	}

	if (pa.Flags&BGPPathAttrFlagOptional) > 0 && pa.Code >= BGPPathAttrTypeUnknown {
		return BGPMessageError{BGPUpdateMsgError, BGPUnrecognizedWellKnownAttr, pkt[:pa.TotalLen()], "Unrecognized Well known attr"}
	}

	return nil
}

func (pa *BGPPathAttrBase) TotalLen() uint32 {
	return uint32(pa.Length) + uint32(pa.BGPPathAttrLen)
}

func (pa *BGPPathAttrBase) GetCode() BGPPathAttrType {
	return pa.Code
}

type BGPPathAttrOrigin struct {
	BGPPathAttrBase
	Value BGPPathAttrOriginType
}

func (o *BGPPathAttrOrigin) Clone() BGPPathAttr {
	x := *o
	x.BGPPathAttrBase = o.BGPPathAttrBase.Clone()
	return &x
}

func (o *BGPPathAttrOrigin) Encode() ([]byte, error) {
	pkt, err := o.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	pkt[o.BGPPathAttrLen] = uint8(o.Value)
	return pkt, nil
}

func (o *BGPPathAttrOrigin) Decode(pkt []byte, data interface{}) error {
	err := o.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	o.Value = BGPPathAttrOriginType(pkt[o.BGPPathAttrLen])

	if o.Value >= BGPPathAttrOriginMax {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, pkt[:o.TotalLen()], fmt.Sprintf("Undefined ORIGIN value %d", uint8(o.Value))}
	}
	return nil
}

func NewBGPPathAttrOrigin(originType BGPPathAttrOriginType) *BGPPathAttrOrigin {
	origin := &BGPPathAttrOrigin{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags:          0x40,
			Code:           BGPPathAttrTypeOrigin,
			Length:         1,
			BGPPathAttrLen: 3,
		},
		Value: originType,
	}

	return origin
}

type BGPASPathSegment interface {
	Clone() BGPASPathSegment
	Encode(pkt []byte) error
	Decode(pkt []byte, data interface{}) error
	PrependAS(as uint32) bool
	TotalLen() uint16
	GetType() BGPASPathSegmentType
	GetLen() uint8
	GetNumASes() uint8
}

type BGPASPathSegmentBase struct {
	Type                BGPASPathSegmentType
	Length              uint8
	BGPASPathSegmentLen uint16
}

func (ps *BGPASPathSegmentBase) Encode(pkt []byte) error {
	pkt[0] = uint8(ps.Type)
	pkt[1] = ps.Length

	return nil
}

func (ps *BGPASPathSegmentBase) Decode(pkt []byte, data interface{}) error {
	if len(pkt) <= 2 {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Not enough data to decode AS path segment"}
	}

	ps.Type = BGPASPathSegmentType(pkt[0])
	ps.Length = pkt[1]

	return nil
}

func (ps *BGPASPathSegmentBase) TotalLen() uint16 {
	return ps.BGPASPathSegmentLen
}

func (ps *BGPASPathSegmentBase) PrependAS(as uint32) bool {
	return true
}

func (ps *BGPASPathSegmentBase) GetType() BGPASPathSegmentType {
	return ps.Type
}

func (ps *BGPASPathSegmentBase) GetLen() uint8 {
	return ps.Length
}

type BGPAS2PathSegment struct {
	BGPASPathSegmentBase
	AS []uint16
}

func (ps *BGPAS2PathSegment) Clone() BGPASPathSegment {
	x := *ps
	x.AS = make([]uint16, len(ps.AS), cap(ps.AS))
	copy(x.AS, ps.AS)
	return &x
}

func (ps *BGPAS2PathSegment) Encode(pkt []byte) error {
	if err := ps.BGPASPathSegmentBase.Encode(pkt); err != nil {
		return err
	}

	for i, as := range ps.AS {
		binary.BigEndian.PutUint16(pkt[(i*2)+2:], as)
	}

	return nil
}

func (ps *BGPAS2PathSegment) Decode(pkt []byte, data interface{}) error {
	if err := ps.BGPASPathSegmentBase.Decode(pkt, data); err != nil {
		return err
	}

	if (len(pkt) - 2) < int(ps.Length)*2 {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Not enough data to decode AS path segment"}
	}

	ps.AS = make([]uint16, ps.Length)
	for i := 0; i < int(ps.Length); i++ {
		ps.AS[i] = binary.BigEndian.Uint16(pkt[(i*2)+2:])
	}
	ps.BGPASPathSegmentLen = uint16(ps.Length*2 + 2)
	return nil
}

func (ps *BGPAS2PathSegment) GetNumASes() uint8 {
	if ps.Type == BGPASPathSet {
		return 1
	} else {
		return uint8(len(ps.AS))
	}
}

func (ps *BGPAS2PathSegment) PrependAS(as uint32) bool {
	if ps.Length >= 255 {
		return false
	}

	ps.AS = append(ps.AS, uint16(as))
	copy(ps.AS[1:], ps.AS[0:])
	ps.AS[0] = uint16(as)
	ps.Length += 1
	ps.BGPASPathSegmentLen += 2
	return true
}

func NewBGPAS2PathSegment(segType BGPASPathSegmentType) *BGPAS2PathSegment {
	as := make([]uint16, 0)
	return &BGPAS2PathSegment{
		BGPASPathSegmentBase: BGPASPathSegmentBase{
			Type:                segType,
			Length:              0,
			BGPASPathSegmentLen: 2,
		},
		AS: as,
	}
}

func NewBGPAS2PathSegmentSeq() *BGPAS2PathSegment {
	return NewBGPAS2PathSegment(BGPASPathSequence)
}

type BGPAS4PathSegment struct {
	BGPASPathSegmentBase
	AS []uint32
}

func (ps *BGPAS4PathSegment) Clone() BGPASPathSegment {
	x := *ps
	x.AS = make([]uint32, len(ps.AS), cap(ps.AS))
	copy(x.AS, ps.AS)
	return &x
}

func (ps *BGPAS4PathSegment) CloneAsAS4PathSegment() *BGPAS4PathSegment {
	x := *ps
	x.AS = make([]uint32, len(ps.AS), cap(ps.AS))
	copy(x.AS, ps.AS)
	return &x
}

func (ps *BGPAS4PathSegment) CloneAsAS2PathSegment() *BGPAS2PathSegment {
	x := NewBGPAS2PathSegment(ps.Type)
	x.AS = make([]uint16, len(ps.AS), cap(ps.AS))
	x.Length = ps.Length
	x.BGPASPathSegmentLen += uint16(x.Length * 2)
	for i, as := range ps.AS {
		if as > math.MaxUint16 {
			x.AS[i] = BGPASTrans
		} else {
			x.AS[i] = uint16(as)
		}
	}
	return x
}

func (ps *BGPAS4PathSegment) Encode(pkt []byte) error {
	if err := ps.BGPASPathSegmentBase.Encode(pkt); err != nil {
		return err
	}

	for i, as := range ps.AS {
		binary.BigEndian.PutUint32(pkt[(i*4)+2:], as)
	}

	return nil
}

func (ps *BGPAS4PathSegment) Decode(pkt []byte, data interface{}) error {
	if err := ps.BGPASPathSegmentBase.Decode(pkt, data); err != nil {
		return err
	}

	if (len(pkt) - 2) < int(ps.Length)*4 {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Not enough data to decode AS path segment"}
	}

	ps.AS = make([]uint32, ps.Length)
	for i := 0; i < int(ps.Length); i++ {
		ps.AS[i] = binary.BigEndian.Uint32(pkt[(i*4)+2:])
	}
	ps.BGPASPathSegmentLen = uint16(ps.Length*4 + 2)
	return nil
}

func (ps *BGPAS4PathSegment) GetNumASes() uint8 {
	if ps.Type == BGPASPathSet {
		return 1
	} else {
		return uint8(len(ps.AS))
	}
}

func (ps *BGPAS4PathSegment) PrependAS(as uint32) bool {
	if ps.Length >= 255 {
		return false
	}

	ps.AS = append(ps.AS, as)
	copy(ps.AS[1:], ps.AS[0:])
	ps.AS[0] = as
	ps.Length += 1
	ps.BGPASPathSegmentLen += 4
	return true
}

func NewBGPAS4PathSegmentSeq() *BGPAS4PathSegment {
	as := make([]uint32, 0)
	return &BGPAS4PathSegment{
		BGPASPathSegmentBase: BGPASPathSegmentBase{
			Type:                BGPASPathSequence,
			Length:              0,
			BGPASPathSegmentLen: 2,
		},
		AS: as,
	}
}

type BGPPathAttrASPath struct {
	BGPPathAttrBase
	Value  []BGPASPathSegment
	ASSize uint8
}

func (as *BGPPathAttrASPath) Clone() BGPPathAttr {
	x := *as
	x.BGPPathAttrBase = as.BGPPathAttrBase.Clone()
	x.Value = make([]BGPASPathSegment, 0, len(as.Value))
	for _, item := range as.Value {
		x.Value = append(x.Value, item.Clone())
	}
	//copy(x.Value, as.Value)
	return &x
}

func (as *BGPPathAttrASPath) CloneAsAS4Path() *BGPPathAttrAS4Path {
	x := NewBGPPathAttrAS4Path()
	x.BGPPathAttrBase = as.BGPPathAttrBase.Clone()
	x.Value = make([]*BGPAS4PathSegment, 0, len(as.Value))
	for _, item := range as.Value {
		x.Value = append(x.Value, item.(*BGPAS4PathSegment).CloneAsAS4PathSegment())
	}
	return x
}

func (as *BGPPathAttrASPath) Encode() ([]byte, error) {
	pkt, err := as.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	length := as.BGPPathAttrBase.BGPPathAttrLen
	for _, val := range as.Value {
		err = val.Encode(pkt[length:])
		if err != nil {
			return pkt, err
		}
		length += val.TotalLen()
	}

	return pkt, nil
}

func (as *BGPPathAttrASPath) Decode(pkt []byte, data interface{}) error {
	err := as.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	as.Value = make([]BGPASPathSegment, 0)
	ptr := uint32(as.BGPPathAttrLen)
	peerAttrs := data.(BGPPeerAttrs)
	as.ASSize = peerAttrs.ASSize
	var asPathSegment BGPASPathSegment
	for ptr < (uint32(as.Length) + uint32(as.BGPPathAttrLen)) {
		if as.ASSize == 4 {
			asPathSegment = NewBGPAS4PathSegmentSeq()
		} else {
			asPathSegment = NewBGPAS2PathSegmentSeq()
		}

		err = asPathSegment.Decode(pkt[ptr:], data)
		if err != nil {
			return nil
		}
		ptr += uint32(asPathSegment.TotalLen())
		if ptr > (uint32(as.Length) + uint32(as.BGPPathAttrLen)) {
			return BGPMessageError{BGPUpdateMsgError, BGPAttrLenError, pkt, "Bad Attribute Length"}
		}
		as.Value = append(as.Value, asPathSegment)
	}
	if ptr != (uint32(as.Length) + uint32(as.BGPPathAttrLen)) {
		return BGPMessageError{BGPUpdateMsgError, BGPAttrLenError, pkt, "Bad Attribute Length"}
	}
	return nil
}

func (as *BGPPathAttrASPath) PrependASPathSegment(pathSeg BGPASPathSegment) {
	as.Value = append(as.Value, pathSeg)
	copy(as.Value[1:], as.Value[0:])
	as.Value[0] = pathSeg
	as.BGPPathAttrBase.Length += pathSeg.TotalLen()
}

func (as *BGPPathAttrASPath) AppendASPathSegment(pathSeg BGPASPathSegment) {
	as.Value = append(as.Value, pathSeg)
	as.BGPPathAttrBase.Length += pathSeg.TotalLen()
}

func NewBGPPathAttrASPath() *BGPPathAttrASPath {
	asPath := &BGPPathAttrASPath{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags: 0x50,
			Code:  BGPPathAttrTypeASPath,
		},
		Value: make([]BGPASPathSegment, 0),
	}
	asPath.BGPPathAttrBase.Length = 0
	asPath.BGPPathAttrBase.BGPPathAttrLen = 4
	return asPath
}

type BGPPathAttrAS4Path struct {
	BGPPathAttrBase
	Value []*BGPAS4PathSegment
}

func (as *BGPPathAttrAS4Path) Clone() BGPPathAttr {
	x := *as
	x.BGPPathAttrBase = as.BGPPathAttrBase.Clone()
	x.Value = make([]*BGPAS4PathSegment, 0, len(as.Value))
	for _, item := range as.Value {
		x.Value = append(x.Value, item.Clone().(*BGPAS4PathSegment))
	}
	//copy(x.Value, as.Value)
	return &x
}

func (as *BGPPathAttrAS4Path) Encode() ([]byte, error) {
	pkt, err := as.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	length := as.BGPPathAttrBase.BGPPathAttrLen
	for _, val := range as.Value {
		err = val.Encode(pkt[length:])
		if err != nil {
			return pkt, err
		}
		length += val.TotalLen()
	}

	return pkt, nil
}

func (as *BGPPathAttrAS4Path) Decode(pkt []byte, data interface{}) error {
	err := as.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	as.Value = make([]*BGPAS4PathSegment, 0)
	ptr := uint32(as.BGPPathAttrLen)
	for ptr < (uint32(as.Length) + uint32(as.BGPPathAttrLen)) {
		asPathSegment := NewBGPAS4PathSegmentSeq()

		err = asPathSegment.Decode(pkt[ptr:], data)
		if err != nil {
			return nil
		}
		ptr += uint32(asPathSegment.TotalLen())
		if ptr > (uint32(as.Length) + uint32(as.BGPPathAttrLen)) {
			return BGPMessageError{BGPUpdateMsgError, BGPAttrLenError, pkt, "Bad Attribute Length"}
		}
		as.Value = append(as.Value, asPathSegment)
	}
	if ptr != (uint32(as.Length) + uint32(as.BGPPathAttrLen)) {
		return BGPMessageError{BGPUpdateMsgError, BGPAttrLenError, pkt, "Bad Attribute Length"}
	}
	return nil
}

func (as *BGPPathAttrAS4Path) AddASPathSegment(pathSeg *BGPAS4PathSegment) {
	as.Value = append(as.Value, pathSeg)
	copy(as.Value[1:], as.Value[0:])
	as.Value[0] = pathSeg
	as.BGPPathAttrBase.Length += pathSeg.TotalLen()
}

func NewBGPPathAttrAS4Path() *BGPPathAttrAS4Path {
	asPath := &BGPPathAttrAS4Path{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags: BGPPathAttrFlagOptional | BGPPathAttrFlagTransitive,
			Code:  BGPPathAttrTypeASPath,
		},
		Value: make([]*BGPAS4PathSegment, 0),
	}
	asPath.BGPPathAttrBase.Length = 0
	asPath.BGPPathAttrBase.BGPPathAttrLen = 4
	return asPath
}

type BGPPathAttrNextHop struct {
	BGPPathAttrBase
	Value net.IP
}

func (n *BGPPathAttrNextHop) Clone() BGPPathAttr {
	x := *n
	x.BGPPathAttrBase = n.BGPPathAttrBase.Clone()
	x.Value = make(net.IP, len(n.Value), cap(n.Value))
	copy(x.Value, n.Value)
	return &x
}

func (n *BGPPathAttrNextHop) Encode() ([]byte, error) {
	pkt, err := n.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	copy(pkt[n.BGPPathAttrBase.BGPPathAttrLen:], n.Value[cap(n.Value)-int(n.Length):])
	return pkt, nil
}

func (n *BGPPathAttrNextHop) Decode(pkt []byte, data interface{}) error {
	err := n.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	n.Value = make(net.IP, n.Length)
	copy(n.Value, pkt[n.BGPPathAttrLen:n.BGPPathAttrLen+n.Length])
	return nil
}

func NewBGPPathAttrNextHop() *BGPPathAttrNextHop {
	return &BGPPathAttrNextHop{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags:          0x40,
			Code:           BGPPathAttrTypeNextHop,
			Length:         4,
			BGPPathAttrLen: 3,
		},
		Value: net.IP{},
	}
}

type BGPPathAttrMultiExitDisc struct {
	BGPPathAttrBase
	Value uint32
}

func (m *BGPPathAttrMultiExitDisc) Clone() BGPPathAttr {
	x := *m
	x.BGPPathAttrBase = m.BGPPathAttrBase.Clone()
	return &x
}

func (m *BGPPathAttrMultiExitDisc) Encode() ([]byte, error) {
	pkt, err := m.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	binary.BigEndian.PutUint32(pkt[m.BGPPathAttrBase.BGPPathAttrLen:], m.Value)
	return pkt, nil
}

func (m *BGPPathAttrMultiExitDisc) Decode(pkt []byte, data interface{}) error {
	err := m.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	m.Value = binary.BigEndian.Uint32(pkt[m.BGPPathAttrLen : m.BGPPathAttrLen+m.Length])
	return nil
}

type BGPPathAttrLocalPref struct {
	BGPPathAttrBase
	Value uint32
}

func (l *BGPPathAttrLocalPref) Clone() BGPPathAttr {
	x := *l
	x.BGPPathAttrBase = l.BGPPathAttrBase.Clone()
	return &x
}

func (l *BGPPathAttrLocalPref) Encode() ([]byte, error) {
	pkt, err := l.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	binary.BigEndian.PutUint32(pkt[l.BGPPathAttrBase.BGPPathAttrLen:], l.Value)
	return pkt, nil
}

func (l *BGPPathAttrLocalPref) Decode(pkt []byte, data interface{}) error {
	err := l.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	l.Value = binary.BigEndian.Uint32(pkt[l.BGPPathAttrLen : l.BGPPathAttrLen+l.Length])
	return nil
}

func NewBGPPathAttrLocalPref() *BGPPathAttrLocalPref {
	return &BGPPathAttrLocalPref{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags:          BGPPathAttrFlagTransitive,
			Code:           BGPPathAttrTypeLocalPref,
			Length:         4,
			BGPPathAttrLen: 3,
		},
	}
}

type BGPPathAttrAtomicAggregate struct {
	BGPPathAttrBase
}

func (a *BGPPathAttrAtomicAggregate) Clone() BGPPathAttr {
	x := *a
	x.BGPPathAttrBase = a.BGPPathAttrBase.Clone()
	return &x
}

type BGPPathAttrAggregator struct {
	BGPPathAttrBase
	AS uint16
	IP net.IP
}

func (a *BGPPathAttrAggregator) Clone() BGPPathAttr {
	x := *a
	x.BGPPathAttrBase = a.BGPPathAttrBase.Clone()
	x.IP = make(net.IP, len(a.IP), cap(a.IP))
	copy(x.IP, a.IP)
	return &x
}

func (a *BGPPathAttrAggregator) Encode() ([]byte, error) {
	pkt, err := a.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	binary.BigEndian.PutUint16(pkt[a.BGPPathAttrBase.BGPPathAttrLen:], a.AS)
	copy(pkt[a.BGPPathAttrBase.BGPPathAttrLen+2:], a.IP)
	return pkt, nil
}

func (a *BGPPathAttrAggregator) Decode(pkt []byte, data interface{}) error {
	err := a.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	a.AS = binary.BigEndian.Uint16(pkt[a.BGPPathAttrLen : a.BGPPathAttrLen+2])
	a.IP = make(net.IP, 4)
	copy(a.IP, pkt[a.BGPPathAttrLen+2:a.BGPPathAttrLen+6])
	return nil
}

func NewBGPPathAttrAggregator() *BGPPathAttrAggregator {
	return &BGPPathAttrAggregator{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags:          BGPPathAttrFlagTransitive | BGPPathAttrFlagOptional,
			Code:           BGPPathAttrTypeAggregator,
			Length:         6,
			BGPPathAttrLen: 3,
		},
		AS: 0,
		IP: net.IP{},
	}
}

type BGPPathAttrAS4Aggregator struct {
	BGPPathAttrBase
	AS uint32
	IP net.IP
}

func (a *BGPPathAttrAS4Aggregator) Clone() BGPPathAttr {
	x := *a
	x.BGPPathAttrBase = a.BGPPathAttrBase.Clone()
	x.IP = make(net.IP, len(a.IP), cap(a.IP))
	copy(x.IP, a.IP)
	return &x
}

func (a *BGPPathAttrAS4Aggregator) Encode() ([]byte, error) {
	pkt, err := a.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	binary.BigEndian.PutUint32(pkt[a.BGPPathAttrBase.BGPPathAttrLen:], a.AS)
	copy(pkt[a.BGPPathAttrBase.BGPPathAttrLen+4:], a.IP)
	return pkt, nil
}

func (a *BGPPathAttrAS4Aggregator) Decode(pkt []byte, data interface{}) error {
	err := a.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	a.AS = binary.BigEndian.Uint32(pkt[a.BGPPathAttrLen : a.BGPPathAttrLen+4])
	a.IP = make(net.IP, 4)
	copy(a.IP, pkt[a.BGPPathAttrLen+4:a.BGPPathAttrLen+8])
	return nil
}

func NewBGPPathAttrAS4Aggregator() *BGPPathAttrAS4Aggregator {
	return &BGPPathAttrAS4Aggregator{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags:          BGPPathAttrFlagTransitive | BGPPathAttrFlagOptional,
			Code:           BGPPathAttrTypeAS4Aggregator,
			Length:         8,
			BGPPathAttrLen: 3,
		},
		AS: 0,
		IP: net.IP{},
	}
}

type BGPPathAttrOriginatorId struct {
	BGPPathAttrBase
	Value net.IP
}

func (o *BGPPathAttrOriginatorId) Clone() BGPPathAttr {
	x := *o
	x.BGPPathAttrBase = o.BGPPathAttrBase.Clone()
	return &x
}

func (o *BGPPathAttrOriginatorId) Encode() ([]byte, error) {
	pkt, err := o.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	copy(pkt[o.BGPPathAttrBase.BGPPathAttrLen:], o.Value)
	return pkt, nil
}

func (o *BGPPathAttrOriginatorId) Decode(pkt []byte, data interface{}) error {
	err := o.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	o.Value = make(net.IP, o.BGPPathAttrBase.Length)
	copy(o.Value, pkt[o.BGPPathAttrLen:o.BGPPathAttrLen+o.BGPPathAttrBase.Length])
	return nil
}

func NewBGPPathAttrOriginatorId(id net.IP) *BGPPathAttrOriginatorId {
	return &BGPPathAttrOriginatorId{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags:          BGPPathAttrFlagOptional,
			Code:           BGPPathAttrTypeOriginatorId,
			Length:         4,
			BGPPathAttrLen: 3,
		},
		Value: id,
	}
}

type BGPPathAttrClusterList struct {
	BGPPathAttrBase
	Value []uint32
}

func (c *BGPPathAttrClusterList) Clone() BGPPathAttr {
	x := *c
	x.BGPPathAttrBase = c.BGPPathAttrBase.Clone()
	x.Value = make([]uint32, len(c.Value))
	for i, _ := range c.Value {
		x.Value[i] = c.Value[i]
	}
	return &x
}

func (c *BGPPathAttrClusterList) Encode() ([]byte, error) {
	pkt, err := c.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, nil
	}

	var i uint16
	for i = 0; i < uint16(len(c.Value)); i++ {
		binary.BigEndian.PutUint32(pkt[c.BGPPathAttrBase.BGPPathAttrLen+(4*i):], c.Value[i])
	}
	return pkt, nil
}

func (c *BGPPathAttrClusterList) Decode(pkt []byte, data interface{}) error {
	pkt, err := c.BGPPathAttrBase.Encode()
	if err != nil {
		return err
	}

	var i uint16
	c.Value = make([]uint32, c.Length/4)
	for i = 0; i < uint16(c.Length/4); i++ {
		c.Value[i] = binary.BigEndian.Uint32(pkt[c.BGPPathAttrLen+(4*i) : c.BGPPathAttrLen+(4*i)+4])
	}
	return nil
}

func (c *BGPPathAttrClusterList) PrependId(id uint32) {
	c.Value = append(c.Value, id)
	copy(c.Value[1:], c.Value[0:])
	c.Value[0] = id
	c.Length += 4
}

func NewBGPPathAttrClusterList() *BGPPathAttrClusterList {
	return &BGPPathAttrClusterList{
		BGPPathAttrBase: BGPPathAttrBase{
			Flags:          BGPPathAttrFlagOptional,
			Code:           BGPPathAttrTypeClusterList,
			Length:         0,
			BGPPathAttrLen: 3,
		},
		Value: make([]uint32, 0),
	}
}

type BGPPathAttrUnknown struct {
	BGPPathAttrBase
	Value []byte
}

func (u *BGPPathAttrUnknown) Clone() BGPPathAttr {
	x := *u
	x.BGPPathAttrBase = u.BGPPathAttrBase.Clone()
	x.Value = make([]byte, len(u.Value), cap(u.Value))
	copy(x.Value, u.Value)
	return &x
}

func (u *BGPPathAttrUnknown) Encode() ([]byte, error) {
	pkt, err := u.BGPPathAttrBase.Encode()
	if err != nil {
		return pkt, err
	}

	copy(pkt[u.BGPPathAttrBase.BGPPathAttrLen:], u.Value)
	return pkt, nil
}

func (u *BGPPathAttrUnknown) Decode(pkt []byte, data interface{}) error {
	err := u.BGPPathAttrBase.Decode(pkt, data)
	if err != nil {
		return err
	}

	u.Value = make([]byte, u.Length)
	copy(u.Value, pkt[u.BGPPathAttrLen:u.BGPPathAttrLen+u.Length])
	return nil
}

func BGPGetPathAttr(pkt []byte) BGPPathAttr {
	typeCode := pkt[1]
	var pathAttr BGPPathAttr

	pathAttr, ok := BGPPathAttrTypeToStructMap[BGPPathAttrType(typeCode)]
	if !ok {
		return &BGPPathAttrUnknown{}
	} else {
		return pathAttr
	}
}

type BGPUpdate struct {
	WithdrawnRoutesLen uint16
	WithdrawnRoutes    []IPPrefix
	TotalPathAttrLen   uint16
	PathAttributes     []BGPPathAttr
	NLRI               []IPPrefix
}

func (msg *BGPUpdate) Clone() BGPBody {
	x := *msg
	x.WithdrawnRoutes = make([]IPPrefix, 0, cap(msg.WithdrawnRoutes))
	for i := 0; i < len(msg.WithdrawnRoutes); i++ {
		//x.WithdrawnRoutes[i] = *msg.WithdrawnRoutes[i].Clone()
		x.WithdrawnRoutes = append(x.WithdrawnRoutes, *msg.WithdrawnRoutes[i].Clone())
	}

	x.PathAttributes = make([]BGPPathAttr, 0, cap(msg.PathAttributes))
	//fmt.Println("BGPUpdate Clone: msg.PathAttrs:", msg.PathAttributes, "len:", len(msg.PathAttributes), "cap:", cap(msg.PathAttributes), "x.PathAttr:", x.PathAttributes, "len:", len(x.PathAttributes), "cap:", cap(x.PathAttributes))
	for i := 0; i < len(msg.PathAttributes); i++ {
		//fmt.Println("BGPUpdate Clone: msg.PathAttrs - i=", i, "attr code=", msg.PathAttributes[i].GetCode())
		//a := msg.PathAttributes[i].Clone()
		//fmt.Println("BGPUpdate Clone:", a)
		//x.PathAttributes[i] = a
		//x.PathAttributes[i] = msg.PathAttributes[i].Clone()
		x.PathAttributes = append(x.PathAttributes, msg.PathAttributes[i].Clone())
	}

	x.NLRI = make([]IPPrefix, 0, cap(msg.NLRI))
	for i := 0; i < len(msg.NLRI); i++ {
		//x.NLRI[i] = *msg.NLRI[i].Clone()
		x.NLRI = append(x.NLRI, *msg.NLRI[i].Clone())
	}
	return &x
}

func (msg *BGPUpdate) Encode() ([]byte, error) {
	pkt := make([]byte, 2)

	for _, route := range msg.WithdrawnRoutes {
		bytes, err := route.Encode()
		if err != nil {
			return pkt, err
		}

		pkt = append(pkt, bytes...)
	}
	wdLen := len(pkt)
	binary.BigEndian.PutUint16(pkt, uint16(wdLen-2))

	pkt = append(pkt, make([]byte, 2)...)
	for _, pa := range msg.PathAttributes {
		bytes, err := pa.Encode()
		if err != nil {
			return pkt, err
		}

		pkt = append(pkt, bytes...)
	}
	paLen := len(pkt) - wdLen
	binary.BigEndian.PutUint16(pkt[wdLen:], uint16(paLen-2))

	for _, nlri := range msg.NLRI {
		bytes, err := nlri.Encode()
		if err != nil {
			return pkt, err
		}

		pkt = append(pkt, bytes...)
	}

	return pkt, nil
}

func (msg *BGPUpdate) decodeIPPrefix(pkt []byte, ipPrefix *[]IPPrefix, length uint32) (uint32, error) {
	ptr := uint32(0)

	if length > uint32(len(pkt)) {
		return ptr, BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Malformed Attributes"}
	}

	for ptr < length {
		ip := IPPrefix{}
		err := ip.Decode(pkt[ptr:])
		if err != nil {
			return ptr, err
		}

		*ipPrefix = append(*ipPrefix, ip)
		ptr += ip.Len()
	}

	if ptr != length {
		return ptr, BGPMessageError{BGPUpdateMsgError, BGPAttrLenError, pkt, "Bad Attribute Length"}
	}
	return ptr, nil
}

func checkPathAttributes(pathAttrs []BGPPathAttr) error {
	found := make(map[BGPPathAttrType]bool)
	for _, attr := range pathAttrs {
		if found[attr.GetCode()] {
			return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil,
				fmt.Sprintf("Path Attr type %d appeared twice in the UPDATE message", attr)}
		}
		found[attr.GetCode()] = true
	}

	for _, attrType := range BGPPathAttrWellKnownMandatory {
		if !found[attrType] {
			return BGPMessageError{BGPUpdateMsgError, BGPMissingWellKnownAttr, []byte{byte(attrType)},
				fmt.Sprintf("Path Attr type %v appeared twice in the UPDATE message", attrType)}
		}
	}

	return nil
}

func (msg *BGPUpdate) Decode(header *BGPHeader, pkt []byte, data interface{}) error {
	msg.WithdrawnRoutesLen = binary.BigEndian.Uint16(pkt[0:2])

	ptr := uint32(2)
	length := uint32(msg.WithdrawnRoutesLen)
	ipLen := uint32(0)
	var err error

	if uint32(msg.WithdrawnRoutesLen)+23 > header.Len() {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Malformed Attributes"}
	}

	msg.WithdrawnRoutes = make([]IPPrefix, 0)
	ipLen, err = msg.decodeIPPrefix(pkt[ptr:], &msg.WithdrawnRoutes, length)
	if err != nil {
		return nil
	}
	ptr += ipLen

	msg.TotalPathAttrLen = binary.BigEndian.Uint16(pkt[ptr : ptr+2])
	ptr += 2

	length = uint32(msg.TotalPathAttrLen)

	if length+uint32(msg.WithdrawnRoutesLen)+23 > header.Len() {
		return BGPMessageError{BGPUpdateMsgError, BGPMalformedAttrList, nil, "Malformed Attributes"}
	}

	msg.PathAttributes = make([]BGPPathAttr, 0)
	for length > 0 {
		pa := BGPGetPathAttr(pkt[ptr:])
		pa.Decode(pkt[ptr:], data)
		msg.PathAttributes = append(msg.PathAttributes, pa)
		ptr += pa.TotalLen()
		length -= pa.TotalLen()
	}

	msg.NLRI = make([]IPPrefix, 0)
	length = header.Len() - 23 - uint32(msg.WithdrawnRoutesLen) - uint32(msg.TotalPathAttrLen)
	ipLen, err = msg.decodeIPPrefix(pkt[ptr:], &msg.NLRI, length)
	if err != nil {
		return nil
	}
	return nil
}

func NewBGPUpdateMessage(wdRoutes []IPPrefix, pa []BGPPathAttr, nlri []IPPrefix) *BGPMessage {
	return &BGPMessage{
		Header: BGPHeader{Type: BGPMsgTypeUpdate},
		Body:   &BGPUpdate{WithdrawnRoutes: wdRoutes, PathAttributes: pa, NLRI: nlri},
	}
}

type BGPMessage struct {
	Header BGPHeader
	Body   BGPBody
}

func (msg *BGPMessage) Clone() *BGPMessage {
	x := *msg
	x.Header = *msg.Header.Clone()
	x.Body = msg.Body.Clone()
	return &x
}

func (msg *BGPMessage) Encode() ([]byte, error) {
	body, err := msg.Body.Encode()
	if err != nil {
		return nil, err
	}

	if msg.Header.Length == 0 {
		if BGPMsgHeaderLen+len(body) > BGPMsgMaxLen {
			return nil, BGPMessageError{0, 0, nil, fmt.Sprintf("BGP message is %d bytes long", BGPMsgHeaderLen+len(body))}
		}
		msg.Header.Length = BGPMsgHeaderLen + uint16(len(body))
	}

	header, err := msg.Header.Encode()
	if err != nil {
		return nil, err
	}
	return append(header, body...), nil
}

func (msg *BGPMessage) Decode(header *BGPHeader, pkt []byte, data interface{}) error {
	msg.Header = *header
	switch header.Type {
	case BGPMsgTypeOpen:
		msg.Body = &BGPOpen{}

	case BGPMsgTypeKeepAlive:
		msg.Body = &BGPKeepAlive{}

	case BGPMsgTypeUpdate:
		msg.Body = &BGPUpdate{}

	case BGPMsgTypeNotification:
		msg.Body = &BGPNotification{}

	default:
		return nil
	}
	err := msg.Body.Decode(header, pkt, data)

	if err == nil && msg.Header.Type == BGPMsgTypeUpdate {
		NormalizeASPath(msg, data)
	}
	return err
}
