
# Storage Layout Reorganization Project

This project aims to reorganize the storage layout of smart contracts while preserving the contract's state. It demonstrates how to handle changes in variable declarations without losing data.

## Getting Started

Follow these steps to set up and run the project:

1. Clone the project
2. Create a directory for test files:
 ```bash
 cd Tests
 mkdir test7
 cd test7
```
3. Create two Solidity files:
```bash
 touch Old.sol
 touch New.sol
```
4. Create two smart contracts in the two files
5. In the New.sol file, you can change the order of declared variables, add new variables, or remove old variables. Ensure that variables in both Old.sol and New.sol with the same names and types are initialized with the same values. If you add new variables, initialize them with 0 or its equivalent for the data type. Note that map data types are not supported yet.
6. Navigate to the Storage_Layout directory and run the following commands to generate the necessary data using the off-chain code analyzer:
```bash
cd ../../Storage_Layout
pipenv shell
python3 main.py
```
7. In the Tests/test7 directory, create two JSON files named old_storage.json and new_storage.json. These files should contain the state of the contract before and after the reorganization, respectively.

## Procedure to Generate State

Go to Remix ide and compile and deploy the contract. After deploying the contract call the compute function and after that press the debug button on the transaction. Then press the "Jump to next breakpoint" button. After that copy the storage.