import asyncio


async def fetch_data():
    await asyncio.sleep(0.1)
    return {"data": 123}


async def async_generator():
    for i in range(3):
        yield i


async def main():
    async for i in async_generator():
        print(i)
    result = await fetch_data()
    print(result)


if __name__ == "__main__":
    asyncio.run(main()) 