// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.2 <0.9.0;

contract MyContract{

    struct Person {
        string name;
        uint age;
    }

    
    Person public person1;
    Person[10] public peopleOfSize10;
    Person[] public peopleDynamic;
    Person[4] public peopleOfSize4;

    function compute() public {

        // Initialize person1
        person1 = Person("John Doe", 30);

        // Add elements to peopleDynamic array
        peopleDynamic.push(Person("Alice", 25));
        peopleDynamic.push(Person("Bob", 32));
        peopleDynamic.push(Person("Venkatanarasimharajuvaripeta Subrahmanyeshwara Rao", 45));

        // Initialize peopleOfSize10 with different names and ages
        peopleOfSize10[0] = Person("Person 1", 20);
        peopleOfSize10[1] = Person("Person 2", 21);
        peopleOfSize10[2] = Person("Person 3", 22);
        peopleOfSize10[3] = Person("Person 4", 23);
        peopleOfSize10[4] = Person("Person 5", 24);
        peopleOfSize10[5] = Person("Person 6", 25);
        peopleOfSize10[6] = Person("Person 7", 26);
        peopleOfSize10[7] = Person("Person 8", 27);
        peopleOfSize10[8] = Person("Person 9", 28);
        peopleOfSize10[9] = Person("Person 10", 29);

        // Initialize peopleOfSize4 with different names and ages
        peopleOfSize4[0] = Person("Person 11", 31);
        peopleOfSize4[1] = Person("Person 12", 32);
        peopleOfSize4[2] = Person("Person 13", 33);
        peopleOfSize4[3] = Person("Person 14", 34);

    }
}