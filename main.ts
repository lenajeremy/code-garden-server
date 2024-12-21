let people = ["Jeremiah", "Lena", "Osaigbokan", "Johnson"]

type Names = typeof people[number]


function isAllowed(n: string): boolean {
    return people.includes(n as any)
}

console.log(people)

console.log(`Jeremiah is included: ${isAllowed("Jeremiah")}`) // check if jeremiah is included
people = people.filter(n => n != "Jeremiah")  // remove jeremiah from the list of people who are included

console.log(people)
console.log(`Jeremiah is included: ${isAllowed("Jeremiah")}`)


console.log(`Johnson is included: ${isAllowed("Johnson")}`)
console.log(`John is included: ${isAllowed("John")}`)