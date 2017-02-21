package validation

import (
	. "GoOnchain/common"
	sig "GoOnchain/core/signature"
	"GoOnchain/crypto"
	. "GoOnchain/errors"
	"GoOnchain/vm"
	"errors"
)

func VerifySignableData(signableData sig.SignableData) error {

	hashes, err := signableData.GetProgramHashes()
	if err != nil {
		return err
	}

	programs := signableData.GetPrograms()
	Length := len(hashes)
	if Length != len(programs) {
		return errors.New("The number of data hashes is different with number of programs.")
	}

	for i := 0; i < Length; i++ {
		hash, err := ToCodeHash(programs[i].Code)
		if err != nil {
			return errors.New("[Validation],VerifySignableData failed.")
		}
		if hashes[i] != hash {
			return errors.New("The data hashes is different with corresponding program code.")
		}
		//execute program on VM
		se := vm.NewExecutionEngine(nil, nil, nil, signableData)
		if se.ExecuteProgram(signableData.GetPrograms()[i].Parameter, false) {
			return NewDetailErr(errors.New("Execute Program Parameter failed."), ErrNoCode, "")
		}
		if se.ExecuteProgram(signableData.GetPrograms()[i].Code, false) {
			return NewDetailErr(errors.New("Execute Program Code failed."), ErrNoCode, "")
		}

		if se.Stack.Count() != 1 || se.Stack.Pop() == nil {
			return NewDetailErr(errors.New("Execute Engine Stack Count Error."), ErrNoCode, "")
		}
	}

	return nil
}

func VerifySignature(signableData sig.SignableData, pubkey *crypto.PubKey, signature []byte) error {
	temp, _ := crypto.Verify(*pubkey, sig.GetHashForSigning(signableData), signature)
	if !temp {
		return NewDetailErr(errors.New("[validation], VerifySignature failed."), ErrNoCode, "")
	} else {
		return nil
	}
}