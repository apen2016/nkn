package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"

	. "github.com/nknorg/nkn/net/protocol"
	"github.com/nknorg/nkn/common/log"
	"github.com/nknorg/nkn/crypto"
	"github.com/nknorg/nkn/common/serialization"
	"github.com/nknorg/nkn/events"
)

type IsingPayload struct {
	PayloadData []byte
	Sender      *crypto.PubKey
	Signature   []byte
}

type IsingMessage struct {
	msgHdr
	pld IsingPayload
}


func (msg IsingMessage) Handle(node Noder) error {
	node.LocalNode().GetEvent("consensus").Notify(events.EventConsensusMsgReceived, &msg.pld)
	return nil
}

func (p *IsingMessage) Serialization() ([]byte, error) {
	msgHeader, err := p.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(msgHeader)
	err = p.pld.Serialize(buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}


func (p *IsingMessage) Deserialization(b []byte) error {
	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.LittleEndian, &(p.msgHdr))
	err = p.pld.Deserialize(buf)

	return err
}

func (p *IsingPayload) SerializeUnsigned(w io.Writer) error {
	err := serialization.WriteVarBytes(w, p.PayloadData)
	if err != nil {
		return err
	}
	err = p.Sender.Serialize(w)
	if err != nil {
		return err
	}

	return nil
}

func (p *IsingPayload) Serialize(w io.Writer) error {
	var err error
	err = p.SerializeUnsigned(w)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, p.Signature)
	if err != nil {
		return err
	}

	return nil
}

func (p *IsingPayload) Deserialize(r io.Reader) error {
	var err error
	err = p.DeserializeUnsigned(r)
	if err != nil {
		return err
	}
	signature, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	p.Signature = signature

	return nil
}

func (p *IsingPayload) DeserializeUnsigned(r io.Reader) error {
	pldData, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	p.PayloadData = pldData

	p.Sender = new(crypto.PubKey)
	err = p.Sender.DeSerialize(r)
	if err != nil {
		return err
	}

	return nil
}

func (p *IsingPayload) DataHash() ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	err := p.SerializeUnsigned(buff)
	if err != nil {
		return nil, err
	}
	temp := sha256.Sum256(buff.Bytes())
	return temp[:], nil
}

func NewIsingConsensus(pld *IsingPayload) ([]byte, error) {
	var msg IsingMessage
	msg.msgHdr.Magic = NETMAGIC
	cmd := "ising"
	copy(msg.msgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer(nil)
	pld.Serialize(tmpBuffer)
	msg.pld = *pld
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}
