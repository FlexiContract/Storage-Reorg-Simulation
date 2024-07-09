// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.2 <0.9.0;

contract MyContract{

    uint64[5][] firstArray;
    uint24[][8] secondArray;
    
     

    function compute() public {
        
        for(uint256 i = 0; i < 3; i++){

            uint64[5] memory tempArray;
            uint64 num = 1;
            for(uint256 j = 0; j < 5; j++){

                tempArray[j] = num;
                num++;
            }
            firstArray.push(tempArray);
        }
        
        for(uint256 i = 0; i < 8; i++){

            

            for(uint24 j = 10; j < 15; j++){

                secondArray[i].push(j);
            }

            
        }
        
            
    }
}