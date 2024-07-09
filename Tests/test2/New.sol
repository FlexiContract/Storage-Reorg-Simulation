// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.2 <0.9.0;

contract MyContract{

    uint64[] secondDynamicArray;
    uint64[4] firstArray;
    uint256[] firstDynamicArray;
    
    

    function compute() public {
        
        for (uint64 i = 0; i < 4; i++){

            firstArray[i] = i+1;
        }

        for (uint256 i = 10; i < 15; i++){

            firstDynamicArray.push(i+1);
        }

        for (uint64 i = 20; i < 25; i++){

            secondDynamicArray.push(i+1);
        }

            
    }
}