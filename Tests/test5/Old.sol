// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.2 <0.9.0;

contract MyContract{

    struct Person {
        string name;
        uint age;
    }

    
    Person public myPerson;

    Person[] public people;

    function compute() public {

        myPerson = Person("John Doe", 30);

        people.push(Person("Alice", 25));
        people.push(Person("Bob", 35));
        people.push(Person("Venkatanarasimharajuvaripeta Subrahmanyeshwara Rao", 40));
    }
}