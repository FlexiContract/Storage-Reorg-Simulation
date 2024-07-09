// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.2 <0.9.0;

contract MyContract{

   uint256 numberOne; 
   string small;
   uint64 numberTwo;
   string big;
    
     

    function compute() public {
        
        numberOne = 7;
        big = "Hello, My name is Tahrim. I am a CS undergrad at University of Dhaka. I really like playing around with blockchain tech.";
        small = "Hello";
        numberTwo = 343;
    }
}