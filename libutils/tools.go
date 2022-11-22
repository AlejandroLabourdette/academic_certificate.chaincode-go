package lib_utils

import (
	"crypto/rand"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"io"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"
)

func UnknownTransactionHandler(ctx contractapi.TransactionContextInterface) error {
	fcn, args := ctx.GetStub().GetFunctionAndParameters()
	return fmt.Errorf("invalid function %s passed with args %v", fcn, args)
}

// GenerateBytesUUID returns a UUID based on RFC 4122 returning the generated bytes
func GenerateBytesUUID() []byte {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		panic(fmt.Sprintf("Error generating UUID: %s", err))
	}

	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80

	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40

	return uuid
}

// GenerateUUID returns a UUID based on RFC 4122
func GenerateUUID() string {
	uuid := GenerateBytesUUID()
	return idBytesToStr(uuid)
}

func idBytesToStr(id []byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:])
}

func RandomNumber(n int) string {
	x1 := mrand.NewSource(time.Now().UnixNano())
	y1 := mrand.New(x1)

	return strconv.Itoa(y1.Intn(n))
}

// KeyResponse contains attribute names and values
type KeyResponse struct {
	ID          string
	CodAsset    string
	YearString  string
	MonthString string
	DayString   string
	TimeString  string
}

// BuildKeyFromID Generate a KeyResponse from an ID
// ID format: COD + YEAR + MONTH + DAY + TIME (hour+minute+second)
// ex: CODE+2022+08+12+103022
func BuildKeyFromID(codAsset, iD string) (*KeyResponse, error) {
	var lCod = 4 // Todos los códigos son de longitud 4

	err := ValidateID(codAsset, iD)
	if err != nil {
		return nil, err
	}

	// TODO: use regexp
	year := iD[lCod : lCod+4]
	month := iD[lCod+4 : lCod+6]
	day := iD[lCod+6 : lCod+8]
	_time := iD[lCod+8 : lCod+14]

	return &KeyResponse{
		ID:          iD,
		CodAsset:    codAsset,
		YearString:  year,
		MonthString: month,
		DayString:   day,
		TimeString:  _time,
		//Consecutive: random,
	}, nil
}

func ValidateID(codAsset, iD string) error {
	var lCod = 4 // Todos los códigos son de longitud 4
	var lID = len(iD)
	var lengthID = lCod + 14

	// check iD length
	if lID != lengthID {
		return fmt.Errorf("invalid id")
	}

	COD := iD[0:lCod]

	if strings.Compare(COD, codAsset) != 0 {
		return fmt.Errorf("invalid id")
	}

	return nil
}

// CompositeKeyFromID
//
// Create composite key from an ID:
// 	ID format: 			COD + YEAR +  MONTH + DAY + TIME
// 	CompositeKey format: objectType +  MONTH + DAY + TIME
//
// 	note: the objectType = COD is assumed
func CompositeKeyFromID(stub shim.ChaincodeStubInterface, objectType string, assetID string) (string, error) {
	responseKey, err := BuildKeyFromID(objectType, assetID)
	if err != nil {
		return "", err
	}

	compositeKey, err := CreateCompositeKeyTo(stub, objectType, responseKey)
	if err != nil {
		return "", err
	} else if compositeKey == "" {
		return "", fmt.Errorf("error creating compound key for: %v", assetID)
	}

	return compositeKey, nil
}

func CreateCompositeKeyTo(stub shim.ChaincodeStubInterface, objectType string, key *KeyResponse) (string, error) {
	return stub.CreateCompositeKey(objectType, []string{key.YearString, key.MonthString, key.DayString, key.TimeString})
}
