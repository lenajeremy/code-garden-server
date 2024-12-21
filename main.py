import os

cwd = os.getcwd()
files_and_folders = os.listdir()
 
for ff in files_and_folders:
    try:
        with open(ff, "r") as f:
            text = f.read()
            print(f"<-----{f.name}------->\n")
            print(text)
            print("\n\n")
    except Exception as e:
        print(f"{ff} is not a file", e)