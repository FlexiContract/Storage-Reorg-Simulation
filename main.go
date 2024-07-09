package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// ANSI escape codes for text colors
	black   = "\033[0;30m"
	red     = "\033[0;31m"
	green   = "\033[0;32m"
	yellow  = "\033[93m"
	orange  = "\033[38;5;208m"
	blue    = "\033[0;34m"
	magenta = "\033[0;35m"
	cyan    = "\033[0;36m"
	white   = "\033[0;37m"
	reset   = "\033[0m"
)

// DummyStateDB to simulate ethereum storage
type DummyStateDB struct {
	Storage map[common.Hash]common.Hash
}

// Helper function to print storage for debugging purpose
func (s *DummyStateDB) PrintStorage(color string) {

	for key, val := range s.Storage {

		fmt.Println(color + fmt.Sprintf("%s : %s", key.Hex(), val.Hex()) + reset)
	}

}

// Helper function to check if storage of two DummyStateDBs are equal. Used for debugging purposes
func (s *DummyStateDB) IsStorageEqual(other *DummyStateDB) error {

	for otherKey, otherVal := range other.Storage {

		if val, found := s.Storage[otherKey]; !found {

			return errors.New("Key Not Found In This State " + otherKey.Hex())

		} else {

			if !bytes.Equal(otherVal[:], val[:]) {

				return errors.New("Mismatch For Key " + otherKey.Hex())
			}
		}
	}

	for key, val := range s.Storage {

		if otherVal, found := other.Storage[key]; !found {

			return errors.New("Key Not Found In Other State " + key.Hex())

		} else {

			if !bytes.Equal(otherVal[:], val[:]) {

				return errors.New("Mismatch For Key " + key.Hex())
			}
		}
	}

	return nil
}

// Gets a storage slot given it's key
func (s *DummyStateDB) GetState(addr common.Address, key common.Hash) common.Hash {

	return s.Storage[key]
}

// Sets a storage slot given the key of the slot and the value to be set
func (s *DummyStateDB) SetState(addr common.Address, key, val common.Hash) {

	if val == (common.Hash{}) {

		delete(s.Storage, key)
	} else {

		s.Storage[key] = val
	}

}

// This is the dummy implementation of a method that we implemented in the go-etehreum source code
// that traverses the storage trie of a given account address and returns the storage slot key value pair
// as a map
func (s *DummyStateDB) GetStorageAsMap(addr common.Address) map[common.Hash]common.Hash {

	return s.Storage
}

// This is the dummy implementation of a method that we implemented in the go-etehreum source code
// that deletes slots from the storage trie of a given account address given the list of keys
func (s *DummyStateDB) DeleteKeysFromStorage(addr common.Address, keys []common.Hash) {

	for _, key := range keys {

		s.SetState(addr, key, common.Hash{})
	}
}

// struct to reoresent a storage slot
type StorageSlot struct {
	Key   common.Hash
	Value common.Hash
}

// struct that holds all the info required to reorganize storage slots
type ReorgInfo struct {
	Type       string      `json:"type"`
	PrevSlot   common.Hash `json:"oldSlot"`
	NewSlot    common.Hash `json:"newSlot"`
	PrevOffset uint64      `json:"oldOffset"`
	NewOffset  uint64      `json:"newOffset"`
}

// struct that holds info of solidity struct type's members
type Member struct {
	PrevOffset uint64      `json:"oldOffset"`
	NewOffset  uint64      `json:"newOffset"`
	PrevSlot   common.Hash `json:"oldSlot"`
	NewSlot    common.Hash `json:"newSlot"`
	Type       string      `json:"type"`
}

// struct to represent data types
type DataType struct {
	Type              string   `json:"type"`
	Base              string   `json:"base"`
	Encoding          string   `json:"encoding"`
	PrevNumberOfBytes uint64   `json:"oldNumberOfBytes"`
	NewNumberOfBytes  uint64   `json:"newNumberOfBytes"`
	Members           []Member `json:"members"`
}

// struct to reorganize storage trie of an ethereum smart contract address
type StorageReorganizer struct {
	state           *DummyStateDB
	commitedStorage map[common.Hash]common.Hash // holds the storage of an account before reorganization
	modifiedStorage map[common.Hash]common.Hash // holds the storage of an account before reorganization
	reorgMessges    []ReorgInfo
	dataTypes       map[string]DataType
	addr            common.Address
}

// Initialization function for the storage reorganizer
func (s *StorageReorganizer) Init(currentState map[common.Hash]common.Hash, reorganizationMessages []ReorgInfo, dataTypes []DataType) {

	s.commitedStorage = currentState
	s.reorgMessges = reorganizationMessages

	for _, dataType := range dataTypes {

		s.dataTypes[dataType.Type] = dataType
	}
}

// function to get commited slot given key
func (s *StorageReorganizer) GetCommitedState(key common.Hash) common.Hash {

	if _, ok := s.commitedStorage[key]; !ok {

		return common.Hash{}
	}

	return s.commitedStorage[key]

}

// function to get modified slot given key
func (s *StorageReorganizer) GetModifiedState(key common.Hash) common.Hash {

	if _, ok := s.modifiedStorage[key]; !ok {

		return common.Hash{}
	}

	return s.modifiedStorage[key]

}

// function to set modified state given key and val
func (s *StorageReorganizer) SetModifiedState(key, val common.Hash) {

	s.modifiedStorage[key] = val
}

// function to check if data type is a struct
func (s *StorageReorganizer) IsStruct(dataType string) (bool, error) {

	if dataType, found := s.dataTypes[dataType]; found {

		if len(dataType.Members) == 0 {

			return false, nil

		} else {

			return true, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

// function to check if a data type is nested. It is considered nested if there is a base or it has members(is a struct)
func (s *StorageReorganizer) IsNested(dataType string) (bool, error) {

	if dataType, found := s.dataTypes[dataType]; found {

		if dataType.Base == "" {

			if len(dataType.Members) == 0 {

				return false, nil
			} else {

				return true, nil
			}

		} else {

			return true, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

// function to check if a data type is flat. It is considered flat
func (s *StorageReorganizer) IsFlat(dataType string) (bool, error) {

	if dataType, found := s.dataTypes[dataType]; found {

		if dataType.Base == "" {

			return true, nil

		} else {

			return false, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

// function to check if the encoding of a data type is "inplace"
func (s *StorageReorganizer) IsEncodingInplace(dataType string) (bool, error) {

	if data, found := s.dataTypes[dataType]; found {

		if data.Encoding == "inplace" {

			return true, nil

		} else {

			return false, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

// function to check if the encoding of a data type is "dynamic_array"
func (s *StorageReorganizer) IsEncodingDynamicArray(dataType string) (bool, error) {

	if data, found := s.dataTypes[dataType]; found {

		if data.Encoding == "dynamic_array" {

			return true, nil

		} else {

			return false, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

// function to check if the encoding of a data type is "bytes"
func (s *StorageReorganizer) IsEncodingBytes(dataType string) (bool, error) {

	if data, found := s.dataTypes[dataType]; found {

		if data.Encoding == "bytes" {

			return true, nil

		} else {

			return false, nil
		}

	} else {

		return false, errors.New("Type not found")
	}
}

// function to get the size of a data type
func (s *StorageReorganizer) GetNumberOfBytes(typeName string) (uint64, uint64, error) {

	if dataType, found := s.dataTypes[typeName]; found {

		return dataType.PrevNumberOfBytes, dataType.NewNumberOfBytes, nil

	} else {

		return 0, 0, errors.New("Type not found")
	}
}

// function to reorganize storage
func (s *StorageReorganizer) Reorganize() error {

	// iterate over the reorg messages
	for _, reorgMessage := range s.reorgMessges {

		// check the encoding of a data type and call functions accordingly
		if isInplace, err := s.IsEncodingInplace(reorgMessage.Type); err != nil {

			return err

		} else if isInplace {
			err := s.ReorganizeInplace(reorgMessage)

			if err != nil {

				return err
			}

		} else if isDynamicArray, err := s.IsEncodingDynamicArray(reorgMessage.Type); err != nil {

			return err

		} else if isDynamicArray {

			err := s.ReorganizeDynamicArray(reorgMessage)

			if err != nil {

				return err
			}

		} else if isBytes, err := s.IsEncodingBytes(reorgMessage.Type); err != nil {

			return err

		} else if isBytes {

			err := s.ReorganizeBytes(reorgMessage)

			if err != nil {

				return err
			}

		} else {

			return errors.New("Not implemented yet")
		}
	}

	return nil

}

// The function iteratively searches through a type's hierarchy to retrieve its type, encoding, and whether it has a non-"inplace" encoding
func (s *StorageReorganizer) ExtractUntilInplace(typeName string) (string, string, bool, error) {

	curType := typeName

	for {

		if dataType, found := s.dataTypes[curType]; found {

			if dataType.Base == "" {

				return dataType.Type, dataType.Encoding, false, nil

			} else {

				if dataType.Encoding != "inplace" {

					return dataType.Type, dataType.Encoding, true, nil

				} else {

					curType = dataType.Base
				}
			}

		} else {

			return "", "", false, errors.New("Type not found")
		}
	}
}

// checks if a data type contains struct inside it
func (s *StorageReorganizer) ContainsStruct(typeName string) (bool, string, error) {

	curType := typeName

	for {
		if dataType, found := s.dataTypes[curType]; found {

			if len(dataType.Members) != 0 {

				return true, curType, nil
			} else {

				if dataType.Base == "" {

					return false, "", nil
				} else {
					curType = dataType.Base
				}
			}
		} else {
			return false, "", errors.New("Type not found")
		}
	}
}

// Reorganizes data type with "inplace" encoding
func (s *StorageReorganizer) ReorganizeInplace(reorgMessage ReorgInfo) error {

	prevNumberOfBytes, _, err := s.GetNumberOfBytes(reorgMessage.Type)

	if err != nil {

		return err
	}

	prevSlotNumber := reorgMessage.PrevSlot.Big()
	newSlotNumber := reorgMessage.NewSlot.Big()

	typeName, encoding, found, err := s.ExtractUntilInplace(reorgMessage.Type)

	if err != nil {

		return err
	}

	structFound, structTypeName, err := s.ContainsStruct(reorgMessage.Type)

	if err != nil {

		return err
	}

	// if there is "dynamic_array" or "bytes" inside the data type then further processing is required
	if found {

		if encoding == "dynamic_array" {

			for i := 0; i < int(prevNumberOfBytes/32); i++ {

				curOldSlotNumber := new(big.Int).Add(new(big.Int).SetInt64(int64(i)), prevSlotNumber)

				curNewSlotNumber := new(big.Int).Add(new(big.Int).SetInt64(int64(i)), newSlotNumber)

				err := s.ReorganizeDynamicArray(ReorgInfo{
					Type:       typeName,
					PrevSlot:   common.BytesToHash(curOldSlotNumber.Bytes()),
					NewSlot:    common.BytesToHash(curNewSlotNumber.Bytes()),
					PrevOffset: 0,
					NewOffset:  0,
				})

				if err != nil {

					return err
				}
			}

		} else if encoding == "bytes" {

			for i := 0; i < int(prevNumberOfBytes/32); i++ {

				curOldSlotNumber := new(big.Int).Add(new(big.Int).SetInt64(int64(i)), prevSlotNumber)

				curNewSlotNumber := new(big.Int).Add(new(big.Int).SetInt64(int64(i)), newSlotNumber)

				err := s.ReorganizeBytes(ReorgInfo{
					Type:       typeName,
					PrevSlot:   common.BytesToHash(curOldSlotNumber.Bytes()),
					NewSlot:    common.BytesToHash(curNewSlotNumber.Bytes()),
					PrevOffset: 0,
					NewOffset:  0,
				})

				if err != nil {

					return err
				}
			}

		} else {

			return errors.New("Not implemented yet")
		}

		return nil

	} else if structFound {
		// if there is a struct inside the inplace data type the members of the struct need to be processed
		prevStructSize, newStructSize, err := s.GetNumberOfBytes(structTypeName)
		if err != nil {
			return err
		}
		structDataType := s.dataTypes[structTypeName]

		curPrevSlot := reorgMessage.PrevSlot.Big()
		curNewSlot := reorgMessage.NewSlot.Big()

		// iterate based on the number of structs
		for i := 0; i < int(prevNumberOfBytes)/int(prevStructSize); i++ {

			//iterate over the members of the struct
			for _, member := range structDataType.Members {

				memberDataType, exists := s.dataTypes[member.Type]
				if !exists {

					return errors.New("Struct Member Not Found")
				}
				//process member according to data type
				if memberDataType.Encoding == "inplace" {
					err := s.ReorganizeInplace(ReorgInfo{
						PrevSlot:   common.BigToHash(new(big.Int).Add(curPrevSlot, member.PrevSlot.Big())),
						NewSlot:    common.BigToHash(new(big.Int).Add(curNewSlot, member.NewSlot.Big())),
						PrevOffset: member.PrevOffset,
						NewOffset:  member.NewOffset,
						Type:       memberDataType.Type,
					})

					if err != nil {

						return err
					}

				} else if memberDataType.Encoding == "dynamic_array" {

					err := s.ReorganizeDynamicArray(ReorgInfo{
						PrevSlot:   common.BigToHash(new(big.Int).Add(curPrevSlot, member.PrevSlot.Big())),
						NewSlot:    common.BigToHash(new(big.Int).Add(curNewSlot, member.NewSlot.Big())),
						PrevOffset: member.PrevOffset,
						NewOffset:  member.NewOffset,
						Type:       memberDataType.Type,
					})

					if err != nil {

						return err
					}

				} else if memberDataType.Encoding == "bytes" {

					err := s.ReorganizeBytes(ReorgInfo{
						PrevSlot:   common.BigToHash(new(big.Int).Add(curPrevSlot, member.PrevSlot.Big())),
						NewSlot:    common.BigToHash(new(big.Int).Add(curNewSlot, member.NewSlot.Big())),
						PrevOffset: member.PrevOffset,
						NewOffset:  member.NewOffset,
						Type:       memberDataType.Type,
					})

					if err != nil {

						return err
					}
				} else {

					return errors.New("Unknown Encoding")
				}
			}

			curPrevSlot = new(big.Int).Add(curPrevSlot, new(big.Int).SetUint64(prevStructSize/32))
			curNewSlot = new(big.Int).Add(curNewSlot, new(big.Int).SetUint64(newStructSize/32))
		}

		return nil

	} else {
		//if the data type does not contain struct or any other type that requires further processing then copy it from the prev slot to the new slot
		var prevOffset, newOffset uint64

		for prevOffset, newOffset = reorgMessage.PrevOffset, reorgMessage.NewOffset; prevOffset < prevNumberOfBytes+reorgMessage.PrevOffset; prevOffset, newOffset = prevOffset+1, newOffset+1 {

			curOldSlotNumber := new(big.Int).Add(new(big.Int).SetUint64(prevOffset/32), prevSlotNumber)

			curNewSlotNumber := new(big.Int).Add(new(big.Int).SetUint64(newOffset/32), newSlotNumber)

			prevSlot := s.GetCommitedState(common.BytesToHash(curOldSlotNumber.Bytes()))
			newSlot := s.GetModifiedState(common.BytesToHash(curNewSlotNumber.Bytes()))

			newSlot[31-(newOffset%32)] = prevSlot[31-(prevOffset%32)]

			s.SetModifiedState(common.BytesToHash(curNewSlotNumber.Bytes()), newSlot)

		}
		return nil
	}

}

func (s *StorageReorganizer) ReorganizeDynamicArray(reorgMessage ReorgInfo) error {

	prevNumberOfBytes, _, err := s.GetNumberOfBytes(reorgMessage.Type)

	if err != nil {

		return err
	}

	//copy the size of the dynamic array from the old slot to new slot
	prevSlot := s.GetCommitedState(reorgMessage.PrevSlot)
	newSlot := s.GetModifiedState(reorgMessage.NewSlot)

	for i := 0; i < int(prevNumberOfBytes); i++ {

		newSlot[i] = prevSlot[i]
	}

	s.SetModifiedState(reorgMessage.NewSlot, newSlot)

	//calculate the slot where data was stored previously and where data will be stored in the reorganized storage structure
	prevDataSlot := common.BytesToHash(crypto.Keccak256(reorgMessage.PrevSlot[:]))
	newDataSlot := common.BytesToHash(crypto.Keccak256(reorgMessage.NewSlot[:]))

	dataType := s.dataTypes[reorgMessage.Type]

	numberOfElements := prevSlot.Big()

	if numberOfElements.Cmp(big.NewInt(0)) == 0 {

		return nil
	}

	//process according to the encoding of the elements of the dynamic array
	if isInplace, err := s.IsEncodingInplace(dataType.Base); err != nil {

		return err

	} else if isInplace {
		// if it is "inplace" then check wether it is flat or nested
		if isNested, err := s.IsNested(dataType.Base); err != nil {

			return err

		} else if isNested {

			prevSizeOfElement, newSizeOfElement, err := s.GetNumberOfBytes(dataType.Base)

			if err != nil {

				return err
			}

			numberOfSlotsPerPrevElement := new(big.Int).SetUint64(prevSizeOfElement / 32)
			numberOfSlotsPerNewElement := new(big.Int).SetUint64(newSizeOfElement / 32)

			for i := big.NewInt(0); i.Cmp(numberOfElements) < 0; i.Add(i, big.NewInt(1)) {

				err := s.ReorganizeInplace(ReorgInfo{
					PrevSlot:   common.BigToHash(new(big.Int).Add(prevDataSlot.Big(), new(big.Int).Mul(numberOfSlotsPerPrevElement, i))),
					NewSlot:    common.BigToHash(new(big.Int).Add(newDataSlot.Big(), new(big.Int).Mul(numberOfSlotsPerNewElement, i))),
					PrevOffset: 0,
					NewOffset:  0,
					Type:       dataType.Base,
				})

				if err != nil {

					return err
				}
			}

		} else if isFlat, err := s.IsFlat(dataType.Base); err != nil {

			return err

		} else if isFlat {

			sizeOfElement, _, err := s.GetNumberOfBytes(dataType.Base)

			if err != nil {

				return err
			}

			numberOfElementsPerSlot := new(big.Int).SetUint64(32 / sizeOfElement)

			numberOfSlots := big.NewInt(0)
			remainder := big.NewInt(0)

			numberOfSlots.DivMod(numberOfElements, numberOfElementsPerSlot, remainder)

			if remainder.Cmp(big.NewInt(0)) > 0 {

				numberOfSlots.Add(numberOfSlots, big.NewInt(1))
			}

			for i := big.NewInt(0); i.Cmp(numberOfSlots) < 0; i.Add(i, big.NewInt(1)) {

				for j := uint64(0); j < 32/sizeOfElement; j++ {

					err := s.ReorganizeInplace(ReorgInfo{
						PrevSlot:   common.BigToHash(new(big.Int).Add(prevDataSlot.Big(), i)),
						NewSlot:    common.BigToHash(new(big.Int).Add(newDataSlot.Big(), i)),
						PrevOffset: j * sizeOfElement,
						NewOffset:  j * sizeOfElement,
						Type:       dataType.Base,
					})

					if err != nil {

						return err
					}
				}
			}

		} else {

			return errors.New("Not Implemented Yet....")
		}

	} else if isDynamicArray, err := s.IsEncodingDynamicArray(dataType.Base); err != nil {

		return err

	} else if isDynamicArray {

		for i := big.NewInt(0); i.Cmp(numberOfElements) < 0; i.Add(i, big.NewInt(1)) {

			err := s.ReorganizeInplace(ReorgInfo{
				PrevSlot:   common.BigToHash(new(big.Int).Add(prevDataSlot.Big(), i)),
				NewSlot:    common.BigToHash(new(big.Int).Add(newDataSlot.Big(), i)),
				PrevOffset: 0,
				NewOffset:  0,
				Type:       dataType.Base,
			})

			if err != nil {

				return err
			}
		}

	} else if isBytes, err := s.IsEncodingBytes(dataType.Base); err != nil {

		return err

	} else if isBytes {

		for i := big.NewInt(0); i.Cmp(numberOfElements) < 0; i.Add(i, big.NewInt(1)) {

			err := s.ReorganizeInplace(ReorgInfo{
				PrevSlot:   common.BigToHash(new(big.Int).Add(prevDataSlot.Big(), i)),
				NewSlot:    common.BigToHash(new(big.Int).Add(newDataSlot.Big(), i)),
				PrevOffset: 0,
				NewOffset:  0,
				Type:       dataType.Base,
			})

			if err != nil {

				return err
			}
		}

	} else {

		return errors.New("Not Implemented Yet....")
	}

	return nil

}

func (s *StorageReorganizer) ReorganizeBytes(reorgMessage ReorgInfo) error {

	numberOfBytes, _, err := s.GetNumberOfBytes(reorgMessage.Type)

	if err != nil {

		return err
	}

	prevSlot := s.GetCommitedState(reorgMessage.PrevSlot)
	newSlot := s.GetModifiedState(reorgMessage.NewSlot)

	//copy data from old slot to new slot
	for i := 0; i < int(numberOfBytes); i++ {

		newSlot[i] = prevSlot[i]
	}

	s.SetModifiedState(reorgMessage.NewSlot, newSlot)

	//calculate the old data slot and new data slot
	prevDataSlot := common.BytesToHash(crypto.Keccak256(reorgMessage.PrevSlot[:]))
	newDataSlot := common.BytesToHash(crypto.Keccak256(reorgMessage.NewSlot[:]))

	// if it is not a short byte do further processing
	if (prevSlot[31] & 1) != 0 {

		numberOfElements := new(big.Int).Div(new(big.Int).Sub(prevSlot.Big(), big.NewInt(1)), big.NewInt(2))
		numberOfSlots := big.NewInt(0)
		remainder := big.NewInt(0)
		numberOfSlots.DivMod(numberOfElements, big.NewInt(32), remainder)

		if remainder.Cmp(big.NewInt(0)) > 0 {

			numberOfSlots.Add(numberOfSlots, big.NewInt(1))
		}

		for i := big.NewInt(0); i.Cmp(numberOfSlots) < 0; i.Add(i, big.NewInt(1)) {

			slotToBeCopiedFrom := common.BigToHash(new(big.Int).Add(prevDataSlot.Big(), i))
			slotToBeCopiedTo := common.BigToHash(new(big.Int).Add(newDataSlot.Big(), i))

			curPrevSlot := s.GetCommitedState(slotToBeCopiedFrom)
			curNewSlot := s.GetModifiedState(slotToBeCopiedTo)

			for j := 0; j < 32; j++ {

				curNewSlot[j] = curPrevSlot[j]
			}

			s.SetModifiedState(slotToBeCopiedTo, curNewSlot)

		}

	}

	return nil

}

// after complete reorganization commit the reorganized state
func (s *StorageReorganizer) Commit() {

	keys := make([]common.Hash, 0)

	for key := range s.commitedStorage {
		keys = append(keys, key)
	}

	s.state.DeleteKeysFromStorage(s.addr, keys)

	for key, val := range s.modifiedStorage {

		if val != (common.Hash{}) {

			s.state.SetState(s.addr, key, val)
		}
	}
}

// returns a new DummyStateDB object initialized with the given state
func NewDummyStateDB(storageSlots *map[common.Hash]StorageSlot) *DummyStateDB {

	storage := make(map[common.Hash]common.Hash)
	for _, slot := range *storageSlots {

		storage[slot.Key] = slot.Value
	}

	return &DummyStateDB{Storage: storage}
}

// returns a new StorageReorganizer object
func NewStorageReorganizer(addr common.Address, state *DummyStateDB) *StorageReorganizer {
	return &StorageReorganizer{
		state:           state,
		commitedStorage: make(map[common.Hash]common.Hash),
		modifiedStorage: make(map[common.Hash]common.Hash),
		dataTypes:       make(map[string]DataType),
		addr:            common.Address{},
	}
}

func ReadStorageFromFile(filePath string) (*map[common.Hash]StorageSlot, error) {
	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return nil, err
	}

	defer file.Close()

	byteVal, _ := ioutil.ReadAll(file)
	var storageSlots map[common.Hash]StorageSlot
	json.Unmarshal(byteVal, &storageSlots)
	for key, slot := range storageSlots {

		if slot.Value.Cmp(common.Hash{}) == 0 {
			delete(storageSlots, key)
		}
	}
	return &storageSlots, nil
}

func ReadReorgInfoFromFile(filePath string) ([]ReorgInfo, error) {

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return nil, err
	}

	defer file.Close()

	byteVal, _ := ioutil.ReadAll(file)
	var reorgInfos []ReorgInfo
	json.Unmarshal(byteVal, &reorgInfos)
	return reorgInfos, nil
}

func ReadDataTypesFromFile(filePath string) ([]DataType, error) {

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return nil, err
	}

	defer file.Close()

	byteVal, _ := ioutil.ReadAll(file)
	var dataTypes []DataType
	json.Unmarshal(byteVal, &dataTypes)
	return dataTypes, nil
}

func getDirectoriesInPath(directoryPath string) ([]string, error) {
	var directories []string

	// Walk the directory and add directory names to the list
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != directoryPath {
			// Exclude the target directory itself
			directories = append(directories, filepath.Base(path))
		}
		return nil
	})

	return directories, err
}

type Result struct {
	directory string
	err       error
}

func runAllTests(targetDirectory string) {

	directories, err := getDirectoriesInPath(targetDirectory)

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return
	}

	var failedTests []Result

	for _, directory := range directories {

		passed, err := runTest(targetDirectory + "/" + directory)

		if passed == false {

			failedTests = append(failedTests, Result{directory: targetDirectory + "/" + directory, err: err})
		}

	}

	if len(failedTests) > 0 {

		fmt.Println(red + "❌❌❌ Failed Tests ❌❌❌" + reset)

		for _, failedTest := range failedTests {

			fmt.Printf("%s%+v%s\n", red, failedTest, reset)
		}

	} else {

		fmt.Println(green + "All passed 🎉🎉🎉" + reset)
	}

}

func runTest(directoryPath string) (bool, error) {

	fmt.Println(cyan + "Current Directory: " + directoryPath + reset)

	storageSlots, err := ReadStorageFromFile(directoryPath + "/" + "old_storage.json")

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return false, err
	}

	dummy := NewDummyStateDB(storageSlots)
	fmt.Println(white + "Before reorganization:" + reset)
	dummy.PrintStorage(yellow)

	reorgInfos, err := ReadReorgInfoFromFile(directoryPath + "/" + "storage_reorg_info.json")

	if err != nil {

		fmt.Println(red + err.Error() + reset)
		return false, err
	}

	/*
		for _, reorgInfo := range reorgInfos {

			fmt.Printf("%+v\n", reorgInfo)

		}
	*/

	dataTypes, err := ReadDataTypesFromFile(directoryPath + "/" + "data_types.json")

	if err != nil {

		fmt.Println(red + err.Error() + reset)
		return false, err
	}

	/*
		for _, nestedType := range nestedTypes {

			fmt.Printf("%+v\n", nestedType)

		}
	*/

	currentStateAsMap := dummy.GetStorageAsMap(common.Address{})
	reorganizer := NewStorageReorganizer(common.Address{}, dummy)
	reorganizer.Init(currentStateAsMap, reorgInfos, dataTypes)

	if reorganizer.Reorganize() != nil {

		return false, err
	}
	reorganizer.Commit()
	fmt.Println(white + "After reorganization:" + reset)
	dummy.PrintStorage(orange)

	expectedStorageSlots, err := ReadStorageFromFile(directoryPath + "/" + "new_storage.json")

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return false, err
	}

	expectedDummy := NewDummyStateDB(expectedStorageSlots)
	err = expectedDummy.IsStorageEqual(dummy)

	if err != nil {
		fmt.Println(red + err.Error() + reset)
		return false, err
	}

	fmt.Println(green + "Test passed: " + directoryPath + "🎉🎉🎉" + reset)
	return true, nil
}

func main() {

	runAllTests("Tests")
	//runTest("Tests/test6")
}
