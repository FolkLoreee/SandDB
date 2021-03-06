var xhr = new XMLHttpRequest();

document.querySelector("#create").addEventListener("submit", handleCreate);
document.querySelector("#read").addEventListener("submit", handleRead);
document.querySelector("#delete").addEventListener("submit", handleDelete);
document.querySelector("#insert").addEventListener("submit", handleInsert);
document.querySelector("#repair").addEventListener("click", handleRepair);
document.querySelector("#full-repair").addEventListener("click", handleFullRepair);

function handleCreate(e) {
    e.preventDefault();
    var result;

    // get input from user
    let table = document.querySelector("#create-table").value;
    let partition_key_names = document.querySelector("#create-partitionkeys").value;
    let clustering_key_names = document.querySelector("#create-clusteringkeys").value;
    if (table == "") {
        alert("Please enter a valid table name");
        return;
    } 
    else if (partition_key_names == "") {
        alert("Please enter valid Partiton Keys");
        return;
    }
    else if (clustering_key_names == "") {
        alert("Please enter valid Clustering Keys");
        return;
    } else {
        partition_key_names = partition_key_names.split(",")
        clustering_key_names = clustering_key_names.split(",")
        var partition_keys = ""
        for (var i = 0; i<partition_key_names.length; i++) {
            if (i == partition_key_names.length-1) {
                partition_keys = partition_keys + `"${partition_key_names[i]}"`    
            } else {
            partition_keys = partition_keys + `"${partition_key_names[i]}",`
            }
        }
        var clustering_keys = ""
        if (clustering_key_names.length === 1) {
            clustering_keys = clustering_keys + `"${clustering_key_names[0]}"`
        } else {
            for (var i = 0; i<clustering_key_names.length; i++) {
                if (i == partition_key_names.length-1) {
                    clustering_keys = clustering_keys + `"${clustering_key_names[i]}"`    
                } else {
                clustering_keys = clustering_keys + `"${clustering_key_names[i]}",`
                }
            }
        }
    }

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
        if (xhr.readyState === 4 && xhr.status === 201) {
            result = xhr.responseText;
            const res = document.querySelector("#create-result");
            res.innerHTML = result;
            return;
        } 
    };

    let data = `{
        "table_name": "${table}",
        "partition_key_names": [${partition_keys}],
        "clustering_key_names": [${clustering_keys}]
    }`;
    console.log(data);

    xhr.send(data);
}

function handleRead(e) {
    e.preventDefault();
    var result;

    // get input from user
    let table = document.querySelector("#read-table").value;
    let partition_keys = document.querySelector("#read-partitionkeys").value;
    let clustering_keys = document.querySelector("#read-clusteringkeys").value;
    // let room = document.querySelector("#read-room").value;
    if (table == "") {
        alert("Please enter a valid table name");
        return;
    } 
    else if (partition_keys == "") {
        alert("Please enter a valid hospital ID");
        return;
    }
    else if (clustering_keys == "") {
        alert("Please enter a valid department");
        return;
    }
    else {
        partition_keys = partition_keys.split(",")
        clustering_keys = clustering_keys.split(",")
        var partitionKeys = ""
        for (var i = 0; i<partition_keys.length; i++) {
            if (i == partition_keys.length-1) {
                partitionKeys = partitionKeys + `"${partition_keys[i]}"`    
            } else {
            partitionKeys = partitionKeys + `"${partition_keys[i]}",`
            }
        }

        var clusteringKeys = ""
        if (clustering_keys.length === 1) {
            clusteringKeys = clusteringKeys + `"${clustering_keys[0]}"`
        } else {
            for (var i = 0; i<clustering_keys.length; i++) {
                if (i == partition_keys.length-1) {
                    clusteringKeys = clusteringKeys + `"${clustering_keys[i]}"`    
                } else {
                clusteringKeys = clusteringKeys + `"${clustering_keys[i]}",`
                }
            }
        }
    }
    

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
            console.log("read successful")
            var response = JSON.parse(xhr.responseText)
            console.log(response["cells"].length)
            result = ""
            if (response["cells"].length === 1) {
                result = result + `${response["cells"][0]["name"]}: ${response["cells"][0]["value"]}`
            } else {
                for (var i = 0; i<response["cells"].length; i++) {
                    result = result + `<p> ${response["cells"][i]["name"]}: ${response["cells"][i]["value"]} </p>`
                }
            }
            console.log(result)
            const res = document.querySelector("#read-result");
            res.innerHTML = result;
        } else {
            result = "error reading table";
            const res = document.querySelector("#read-result");
            res.innerHTML = result;
        }
    };

    let data = `{
        "table_name": "${table}",
        "partition_keys": [${partitionKeys}],
        "clustering_keys": [${clusteringKeys}]
    }`;
    console.log(data)

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
        "table_name": "${table}",
        "partition_key_names": ["${hospitalId}","${department}"],
        "clustering_keys_names": ["${room}"]
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
        resourceName = resourceName.split(",")
        resourceValue = resourceValue.split(",")
        var resource_names = ""
        for (var i = 0; i<resourceName.length; i++) {
            if (i == resourceName.length-1) {
                resource_names = resource_names + `"${resourceName[i]}"`    
            } else {
            resource_names = resource_names + `"${resourceName[i]}",`
            }
        }
        var resource_values = ""
        for (var i = 0; i<resourceValue.length; i++) {
            if (i == resourceValue.length-1) {
                resource_values = resource_values + `"${resourceValue[i]}"`    
            } else {
            resource_values = resource_values + `"${resourceValue[i]}",`
            }
        }
    }

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
            result = "data successfully written";
            const res = document.querySelector("#insert-result");
            res.innerHTML = result;
        } 
        else {
            console.log(xhr.status)
            result = "error writing data";
            const res = document.querySelector("#insert-result");
            res.innerHTML = result;
        }
    };

    let data = `{
        "table_name": "${table}",
        "partition_keys": ["${hospitalId}","${department}"],
        "clustering_keys": ["${room}"],
        "cell_names": [${resource_names}],
        "cell_values": [${resource_values}]
    }`;

    xhr.send(data);
}

function handleRepair(e) {
    e.preventDefault();

    let url = "http://127.0.0.1:8000/repair";
    console.log(url);
    xhr.open("POST", url);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status == 200) {
            const res = document.querySelector("#repair-result");
            res.innerHTML = xhr.statusText;
        } else {
            const res = document.querySelector("#repair-result");
            res.innerHTML = xhr.statusText;
            return;
        }
    };

    let data = `{
    }`;

    xhr.send(data);
}

function handleFullRepair(e) {
    e.preventDefault();

    let url = "http://127.0.0.1:8000/full_repair";
    console.log(url);
    xhr.open("POST", url);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status == 200) {
            const res = document.querySelector("#full-repair-result");
            res.innerHTML = xhr.statusText;
        } else {
            alert("server issue");
            return;
        }
    };

    let data = `{
    }`;

    xhr.send(data);
}