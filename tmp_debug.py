import json

with open("empty_operation_hours.txt", "r", encoding="utf-8") as f:
    text = f.read()

for line in text.split("\n"):
    if line.startswith('    "d": '):
        v = line[9:-2]
        # v is a JSON literal string representing a string
        json_str = json.loads('"' + v + '"')
        # There's )]}'\n at the start of json_str
        json_str = json_str.split('\n', 1)[1]
        data = json.loads(json_str)
        
        darray = data[0][1]
        try:
            print("203:", len(darray[203][0]) if (len(darray) > 203 and darray[203]) else "None")
        except Exception as e:
            print("203 err", e)
            
        try:
            print("34:", len(darray[34][1]) if (len(darray) > 34 and darray[34] and len(darray[34]) > 1) else "None")
        except Exception as e:
            print("34 err", e)
