from typing import List, Dict, Union


def process(items: List[int] | tuple[int, ...]) -> Dict[str, Union[int, float]]:
    return {"total": sum(items), "average": sum(items) / len(items)}


if __name__ == "__main__":
    print(process([1, 2, 3])) 