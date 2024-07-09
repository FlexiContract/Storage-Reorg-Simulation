import subprocess
import json
import sys
import os
import re

def int_to_256bit_hex_string(num):
    # Convert the integer to a hex string
    hex_string = hex(num)[2:]

    # Ensure the hex string is 256 bits long
    padded_hex_string = hex_string.zfill(64)

    # Add the "0x" prefix
    final_hex_string = "0x" + padded_hex_string

    return final_hex_string

def get_storage_layout(file_name):
    # Command to run (solc --stage-layout sample.sol)
    command = ["solc", "--storage-layout"]
    command.append(file_name)

    # Run the command
    result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, universal_newlines=True)

    # Check if the command was successful
    if result.returncode == 0:
        output = result.stdout

        # Find the index of "Contract Storage Layout:"
        layout_start_index = output.find("Contract Storage Layout:")

        # Extract the data after "Contract Storage Layout:"
        if layout_start_index != -1:
            layout_data_str = output[layout_start_index + len("Contract Storage Layout:"):].strip()

            # Convert the string to a dictionary using json.loads
            try:
                layout_data_dict = json.loads(layout_data_str)
                return layout_data_dict
            except json.JSONDecodeError as e:
                raise Exception("Error decoding JSON")
        else:
            raise Exception("Contract Storage Layout not found in the output.")
    else:
        raise Exception(result.stderr)

#check if struct is present inside type
def is_struct_present(type_id,types):
    type = types[type_id]
    
    if "members" in type:
        return True
    
    if "base" in type:
        return is_struct_present(type["base"],types)
    else:
        return False    

#check if two types are equal
def is_type_equal(old_type_id,new_type_id,old_types,new_types):

    #check if type names are equal
    if old_type_id != new_type_id:
        return False
    
    old_type = old_types[old_type_id]
    new_type = new_types[new_type_id]

    #check if there is struct inside the data type
    is_struct_present_in_old_type = is_struct_present(old_type_id,old_types)
    is_struct_present_in_new_type = is_struct_present(new_type_id,new_types)

    #if there is no struct inside either of them then check if they contain equal number of bytes
    if is_struct_present_in_new_type == False and is_struct_present_in_old_type == False:
        if old_type["numberOfBytes"] != new_type["numberOfBytes"]:
            return False

    #check if both data types have same label and encoding
    if old_type["label"] != new_type["label"] or old_type["encoding"] != new_type["encoding"]:
        return False

    #if one of the data types has a key named base and the other one does not they are not same
    if ("base" in old_type and "base" not in new_type) or ("base" not in old_type and "base" in new_type):
        return False
    
    #if both have the base key check if the base type is equal
    if "base" in old_type and "base" in new_type:
        res = is_type_equal(old_type["base"],new_type["base"],old_types,new_types)
        if res == False:
            return False
    
    #if one of the data types has a key named members and the other one does not they are not same
    if ("members" in old_type and "members" not in new_type) or ("members" not in old_type and "members" in new_type):
        return False

    #if both have the members key check if the members type is equal
    if "members" in old_type and "members" in new_type:
        if len(old_type["members"]) == 0 or len(new_type["members"]) == 0:
            return False
        
        old_members = {}
        new_members = {}
        
        for member_in_old in old_type["members"]:
            old_members[member_in_old["label"]] = member_in_old

        for member_in_new in new_type["members"]:
            new_members[member_in_new["label"]] = member_in_new

        #check for matching members
        found_matching_members = False
        for key,val in old_members.items():
            if key not in new_members:
                continue
            new_val = new_members[key]
            res = is_type_equal(val["type"],new_val["type"],old_types,new_types)
            if res == True:
                found_matching_members = True

        if found_matching_members == False:
            return False    

    return True

def get_objects(old_json, new_json):
    old_storage = old_json["storage"] #storage in the old contract
    new_storage = new_json["storage"] #storage in the new contract
    old_types = old_json["types"] #data types in the old contract
    new_types = new_json["types"] #data types in the new contract
    
    common_objects = [] #list to hold storage objects both in old and new contract
    
    for old_storage_object in old_storage:
        for new_storage_object in new_storage:
            #if the storage objects from the old and the new contract have the same label and their data types are the same then insert into common objects list
            if old_storage_object["label"] == new_storage_object["label"] and is_type_equal(old_storage_object["type"],new_storage_object["type"],old_types,new_types) == True:
                    common_objects.append({
                        "label":old_storage_object["label"],
                        "type":old_storage_object["type"],
                        "oldSlot":int_to_256bit_hex_string(int(old_storage_object["slot"])),
                        "newSlot":int_to_256bit_hex_string(int(new_storage_object["slot"])),
                        "oldOffset":old_storage_object["offset"],
                        "newOffset":new_storage_object["offset"],                       
                    })
    return common_objects

"""
def get_types(old_json, common_objects):
    old_types = old_json["types"]
    inserted_types = []
    nested_types = []
    flat_types = []
    for common_object in common_objects:
        current_type = common_object["type"]
        while old_types.get(current_type,None) is not None and current_type not in inserted_types:
            old_types[current_type]["type"] = current_type
            old_types[current_type]["numberOfBytes"] = int(old_types[current_type]["numberOfBytes"])
            inserted_types.append(current_type)
            if "base" in old_types[current_type]:
                nested_types.append(old_types[current_type])
                current_type = old_types[current_type]["base"]
            else:
                flat_types.append(old_types[current_type])
                break
        if old_types.get(current_type,None) is None:
            raise Exception("Type not found....")
    return nested_types,flat_types
"""
def process_type(old_types, new_types, current_type, inserted_types, data_types):
    
    if current_type in inserted_types:
        return
    
    old_types[current_type]["type"] = current_type #add a key value pair named type
    old_types[current_type]["oldNumberOfBytes"] = int(old_types[current_type]["numberOfBytes"]) #add key value pair that contains the size of the data type in the old contract
    old_types[current_type]["newNumberOfBytes"] = int(new_types[current_type]["numberOfBytes"]) #add key value pair that contains the size of the data type in the new contract

    inserted_types.append(current_type)
    #if there is a base type process it too
    if "base" in old_types[current_type]:
        process_type(old_types,new_types,old_types[current_type]["base"],inserted_types,data_types)
    else:
        old_types[current_type]["base"] = None

    #if the data type is a struct then process the members
    if "members" in old_types[current_type]:
        old_members = {}
        new_members = {}
        
        for member_in_old in old_types[current_type]["members"]:
            old_members[member_in_old["label"]] = member_in_old

        for member_in_new in new_types[current_type]["members"]:
            new_members[member_in_new["label"]] = member_in_new
        
        for member in old_types[current_type]["members"]:
            if member["label"] not in new_members:
                old_types[current_type]["members"].remove(member)
                continue
            member_in_new = new_members[member["label"]]
            member["oldSlot"] = member["slot"]
            member["newSlot"] = member_in_new["slot"]
            member["oldOffset"] = member["offset"]
            member["newOffset"] = member_in_new["offset"]
            process_type(old_types,new_types,member["type"],inserted_types,data_types)
    else:
        old_types[current_type]["members"] = None


    data_types.append(old_types[current_type])    

#find the data types of the storage objects that require reorganization
def get_types(old_types, new_types, common_objects):
    inserted_types = []
    data_types = []
    
    for common_object in common_objects:
        current_type = common_object["type"]
        process_type(old_types,new_types,current_type,inserted_types,data_types)
        if old_types.get(current_type,None) is None:
            raise Exception("Type not found....")
    
    for type in data_types:
        if type["members"] is not None:
            for member in type["members"]:
                member["slot"] = int_to_256bit_hex_string(int(member["slot"]))
                member["oldSlot"] = int_to_256bit_hex_string(int(member["oldSlot"]))
                member["newSlot"] = int_to_256bit_hex_string(int(member["newSlot"]))
                member.pop("astId")
                member.pop("contract")
                member.pop("label")
                member.pop("slot")
    return data_types

def writeJSON(file_name,data):
    with open(file_name, 'w') as json_file:
        json.dump(data, json_file, indent=2)


def get_directories_in_path(directory_path):
    # Get all files and directories in the specified path
    entries = os.listdir(directory_path)

    # Filter out directories
    directories = [entry for entry in entries if os.path.isdir(os.path.join(directory_path, entry))]

    return directories

#modify struct type name
def modify_struct_types(text):
    pattern = r't_struct\((.*?)\)[a-zA-Z0-9]+_storage'
    result = re.sub(pattern, r't_struct(\1)_storage', text)
    return result

def clean_types(storage_layout):
    storage = storage_layout["storage"]
    types = storage_layout["types"]

    for item in storage:
        modified_type_def = modify_struct_types(item["type"])
        item["type"] = modified_type_def

    all_keys = list(types.keys())
    for key in all_keys:
        new_key = modify_struct_types(key)
        if "base" in types[key]:
            base = types[key]["base"]
            new_base = modify_struct_types(base)
            if new_base != base:
                types[key]["base"] = new_base
        if new_key != key:
            types[new_key] = types.pop(key)

    storage_layout["storage"] = storage
    storage_layout["types"] = types



if __name__ == "__main__":
    
    target_directory = "../Tests"
    directories_list = get_directories_in_path(target_directory)

    # Print the result
    for directory in directories_list:
        current_directory = target_directory+"/"+directory
        old_file = current_directory+"/"+"Old.sol"
        new_file = current_directory+"/"+"New.sol"
        old_storage_layout = get_storage_layout(old_file)
        clean_types(old_storage_layout)
        new_storage_layout = get_storage_layout(new_file)
        clean_types(new_storage_layout)
        
        result = get_objects(old_storage_layout, new_storage_layout)
        #nested,flat = get_types(old_storage_layout,result)
        data_types = get_types(old_storage_layout["types"],new_storage_layout["types"],result)
        #print(json.dumps(data_types,indent=2))
        
        writeJSON(current_directory+"/"+"storage_reorg_info.json",result)
        #writeJSON(current_directory+"/"+"nested_types.json",nested)
        #writeJSON(current_directory+"/"+"flat_types.json",flat)
        writeJSON(current_directory+"/"+"data_types.json",data_types)
        