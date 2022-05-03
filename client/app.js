var xhr = new XMLHttpRequest();

document.querySelector("#create").addEventListener("submit", handleCreate);
document.querySelector("#read").addEventListener("submit", handleRead);
document.querySelector("#delete").addEventListener("submit", handleDelete);
document.querySelector("#insert").addEventListener("submit", handleInsert);

function handleCreate(e) {
    e.preventDefault();
    var result;

    // get input from user
    let table = document.querySelector("#create-table").value;
    let hospitalId = document.querySelector("#create-hospitalID").value;
    let department = document.querySelector("#create-department").value;
    let room = document.querySelector("#create-room").value;
    if (table == "") {
        alert("Please enter a valid table name");
        return;
    } 
    else if (hospitalId == "") {
        alert("Please enter a valid hospital ID");
        return;
    }
    else if (department == "") {
        alert("Please enter a valid department");
        return;
    }
    else if (room == "") {
        alert("Please enter a valid room ID");
        return;
    }
    else {
        department = department.toUpperCase();
        room = room.toUpperCase();
    }
    
    // result = table+hospitalId+department+room
    // const res = document.querySelector("#create-result");
    // res.innerHTML = result;

    // send input to endpoints
    const ports = ["8000", "8001", "8002", "8003"];
    var portId = Math.floor(Math.random() * 10);
    if (portId > 3) {
        portId %= 4;
    };
    const port = ports[portId];
    let url = "http://127.0.0.1:" + port + "/create";
    console.log(url);
    xhr.open("POST", url);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status == 200) {
            result = xhr.responseText;
            const res = document.querySelector("#create-result");
            res.innerHTML = result;
        } else {
            alert("server issue");
            return;
        }
    };

    let data = `{
        "table_name": "${table}",
        "partition_key_names": ["HOSPITAL_ID_${hospitalId}","${department}_DEPT"],
        "clustering_key_names": ["${room}"]
    }`;
    xhr.send(data);
}

function handleRead(e) {
    e.preventDefault();
    var result;

    // get input from user
    let table = document.querySelector("#read-table").value;
    let hospitalId = document.querySelector("#read-hospitalID").value;
    let department = document.querySelector("#read-department").value;
    let room = document.querySelector("#read-room").value;
    if (table == "") {
        alert("Please enter a valid table name");
        return;
    } 
    else if (hospitalId == "") {
        alert("Please enter a valid hospital ID");
        return;
    }
    else if (department == "") {
        alert("Please enter a valid department");
        return;
    }
    else if (room == "") {
        alert("Please enter a valid room ID");
        return;
    }
    else {
        department = department.toUpperCase();
        room = room.toUpperCase();
    }
    
    // const res = document.querySelector("#general-result");
    // res.innerHTML = result;

    // send input to endpoints
    const ports = ["8000", "8001", "8002", "8003"];
    var portId = Math.floor(Math.random() * 10);
    if (portId > 3) {
        portId %= 4;
    };
    const port = ports[portId];
    let url = "http://127.0.0.1:" + port + "/read";
    console.log(url);
    xhr.open("POST", url);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status == 200) {
            result = xhr.responseText;
            const res = document.querySelector("#read-result");
            res.innerHTML = result;
        } else {
            alert("server issue");
            return;
        }
    };

    let data = `{
        "table_name": "${table}",
        "partition_keys": ["HOSPITAL_ID_${hospitalId}","${department}_DEPT"],
        "clustering_keys": ["${room}"]
    }`;
    console.log(data);
    xhr.send(data);
}

function handleDelete(e) {
    e.preventDefault();
    var result;

    // get input from user
    let table = document.querySelector("#delete-table").value;
    let hospitalId = document.querySelector("#delete-hospitalID").value;
    let department = document.querySelector("#delete-department").value;
    let room = document.querySelector("#delete-room").value;
    if (table == "") {
        alert("Please enter a valid table name");
        return;
    } 
    else if (hospitalId == "") {
        alert("Please enter a valid hospital ID");
        return;
    }
    else if (department == "") {
        alert("Please enter a valid department");
        return;
    }
    else if (room == "") {
        alert("Please enter a valid room ID");
        return;
    }
    else {
        department = department.toUpperCase();
        room = room.toUpperCase();
    }
    
    // const res = document.querySelector("#general-result");
    // res.innerHTML = result;

    // send input to endpoints
    const ports = ["8000", "8001", "8002", "8003"];
    var portId = Math.floor(Math.random() * 10);
    if (portId > 3) {
        portId %= 4;
    };
    const port = ports[portId];
    let url = "http://127.0.0.1:" + port + "/delete";
    console.log(url);
    xhr.open("POST", url);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status == 200) {
            result = xhr.responseText;
            const res = document.querySelector("#delete-result");
            res.innerHTML = result;
        } else {
            alert("server issue");
            return;
        }
    };

    let data = `{
        table_name: ${table},
        partition_key_names: ["HOSPITAL_ID_${hospitalId}","${department}_DEPT"],
        clustering_key_names: ["${room}"],
    }`;
    xhr.send(data);
}

function handleInsert(e) {
    e.preventDefault();
    var result;

    // get input from user
    let table = document.querySelector("#insert-table").value;
    let hospitalId = document.querySelector("#insert-hospitalID").value;
    let department = document.querySelector("#insert-department").value;
    let room = document.querySelector("#insert-room").value;
    let resourceName = document.querySelector("#resource-name").value
    let resourceValue = document.querySelector("#resource-value").value
    if (table == "") {
        alert("Please enter a valid table name");
        return;
    } 
    else if (hospitalId == "") {
        alert("Please enter a valid hospital ID");
        return;
    }
    else if (department == "") {
        alert("Please enter a valid department");
        return;
    }
    else if (room == "") {
        alert("Please enter a valid room ID");
        return;
    }
    else if (resourceName == "") {
        alert("Please enter a valid resource name")
        return 
    }
    else if (resourceValue == "") {
        alert("Please enter a valid resource amount")
        return
    }
    else {
        department = department.toUpperCase();
        room = room.toUpperCase();
    }
    
    // const res = document.querySelector("#insert-result");
    // res.innerHTML = result;

    // send input to endpoints
    const ports = ["8000", "8001", "8002", "8003"];
    var portId = Math.floor(Math.random() * 10);
    if (portId > 3) {
        portId %= 4;
    };
    const port = ports[portId];
    let url = "http://127.0.0.1:" + port + "/insert";
    console.log(url);
    xhr.open("POST", url);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status == 200) {
            result = xhr.responseText;
            const res = document.querySelector("#insert-result");
            res.innerHTML = result;
        } else {
            alert("server issue");
            return;
        }
    };

    let data = `{
        "table_name": "${table}",
        "partition_keys": ["HOSPITAL_ID_${hospitalId}","${department}_DEPT"],
        "clustering_keys": ["${room}"],
        "cell_names": ["${resourceName}"],
        "cell_values": ["${resourceValue}"]
    }`;
    console.log(data);
    xhr.send(data);
}

function getInputValues() {
    

    return table, hospitalId, department, room;
}
