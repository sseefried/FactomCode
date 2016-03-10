// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
	"io"
)

// defaultInvListAlloc is the default size used for the backing array for an
// inventory list.  The array will dynamically grow as needed, but this
// figure is intended to provide enough space for the max number of inventory
// vectors in a *typical* inventory message without needing to grow the backing
// array multiple times.  Technically, the list can grow to MaxInvPerMsg, but
// rather than using that large figure, this figure more accurately reflects the
// typical case.
//const defaultInvListAlloc = 1000

// MsgDirInv implements the Message interface and represents a bitcoin inv message.
// It is used to advertise a peer's known data such as blocks and transactions
// through inventory vectors.  It may be sent unsolicited to inform other peers
// of the data or in response to a getDirblocks message (MsgGetDirBlocks).  Each
// message is limited to a maximum number of inventory vectors, which is
// currently 50,000.
//
// Use the AddInvVect function to build up the list of inventory vectors when
// sending an inv message to another peer.
type MsgDirInv struct {
	InvList []*InvVect
}

// AddInvVect adds an inventory vector to the message.
func (msg *MsgDirInv) AddInvVect(iv *InvVect) error {
	if len(msg.InvList)+1 > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [max %v]",
			MaxInvPerMsg)
		return messageError("MsgDirInv.AddInvVect", str)
	}

	msg.InvList = append(msg.InvList, iv)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgDirInv) BtcDecode(r io.Reader, pver uint32) error {
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}

	// Limit to max inventory vectors per message.
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgDirInv.BtcDecode", str)
	}

	msg.InvList = make([]*InvVect, 0, count)
	for i := uint64(0); i < count; i++ {
		iv := InvVect{}
		err := readInvVect(r, pver, &iv)
		if err != nil {
			return err
		}
		msg.AddInvVect(&iv)
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgDirInv) BtcEncode(w io.Writer, pver uint32) error {
	// Limit to max inventory vectors per message.
	count := len(msg.InvList)
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgDirInv.BtcEncode", str)
	}

	err := writeVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, iv := range msg.InvList {
		err := writeInvVect(w, pver, iv)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgDirInv) Command() string {
	return CmdDirInv
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgDirInv) MaxPayloadLength(pver uint32) uint32 {
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	return MaxVarIntPayload + (MaxInvPerMsg * maxInvVectPayload)
}

// NewMsgDirInv returns a new bitcoin inv message that conforms to the Message
// interface.  See MsgDirInv for details.
func NewMsgDirInv() *MsgDirInv {
	return &MsgDirInv{
		InvList: make([]*InvVect, 0, defaultInvListAlloc),
	}
}

// NewMsgDirInvSizeHint returns a new bitcoin inv message that conforms to the
// Message interface.  See MsgDirInv for details.  This function differs from
// NewMsgDirInv in that it allows a default allocation size for the backing array
// which houses the inventory vector list.  This allows callers who know in
// advance how large the inventory list will grow to avoid the overhead of
// growing the internal backing array several times when appending large amounts
// of inventory vectors with AddInvVect.  Note that the specified hint is just
// that - a hint that is used for the default allocation size.  Adding more
// (or less) inventory vectors will still work properly.  The size hint is
// limited to MaxInvPerMsg.
func NewMsgDirInvSizeHint(sizeHint uint) *MsgDirInv {
	// Limit the specified hint to the maximum allow per message.
	if sizeHint > MaxInvPerMsg {
		sizeHint = MaxInvPerMsg
	}

	return &MsgDirInv{
		InvList: make([]*InvVect, 0, sizeHint),
	}
}