// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.2 <0.9.0;

contract MyContract{

    struct Person {
        string name;
        uint income;
        uint age;
    }

    
    
    Person[4] public peopleOfSize4;
    Person public person1;
    Person[] public peopleDynamic;
    

    function compute() public {

        // Initialize person1
        person1 = Person("John Doe", 50000, 30);

        // Add elements to peopleDynamic array
        peopleDynamic.push(Person("Alice", 60000, 25));
        peopleDynamic.push(Person("Bob", 55000, 32));
        peopleDynamic.push(Person("Venkatanarasimharajuvaripeta Subrahmanyeshwara Rao", 70000, 45));

        // Initialize peopleOfSize4 with different names, incomes, and ages
        peopleOfSize4[0] = Person("Person 11", 110000, 31);
        peopleOfSize4[1] = Person("Person 12", 120000, 32);
        peopleOfSize4[2] = Person("Person 13", 130000, 33);
        peopleOfSize4[3] = Person("Person 14", 140000, 34);

    }
}