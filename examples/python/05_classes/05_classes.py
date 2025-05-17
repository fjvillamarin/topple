class Animal:
    kingdom = "Animalia"

    def __init__(self, name: str):
        self.name = name

    def speak(self) -> str:
        return f"{self.name} makes a sound"


class Dog(Animal):
    def speak(self) -> str:
        return f"{self.name} says woof!"


if __name__ == "__main__":
    dog = Dog("Fido")
    print(dog.speak()) 