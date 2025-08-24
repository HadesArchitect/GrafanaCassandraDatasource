# Quick Demo

If you want to try the data source in action, please follow the easy steps of this quick demo. This demo has a simple docker-compose setup with three services running:

* Apache Cassandra 4.0
* [Sample Data](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/demo/sample_data.sh) (a simple bash script generating sample "temperature" data from some "sensors")
* Grafana 10 with preinstalled Cassandra Datasource 

# Prerequisites

* `docker`
* `docker-compose`

## Quick Launch

### 1. Download the repository using `git clone` (a simple archive download will work too)

```
git clone https://github.com/HadesArchitect/GrafanaCassandraDatasource.git 
cd GrafanaCassandraDatasource/demo
```

**IMPORTANT** Switch to the `GrafanaCassandraDatasource/demo` folder as in the instruction above (notice the `demo` subfolder).

### 2. Launch the project

```
docker-compose up -d
```

### 3. Wait for Cassandra to start, it may take a bit

Check the status with the command:

```
docker-compose exec cassandra nodetool status | grep rack1
```

As soon as the output looks like `UN  172.23.0.3 88.16 KiB ...` (UN means UP and NORMAL), you are good to go.

### 4. Login to Grafana

Open `http://localhost:3000/dashboards` in your browser. Choose the `Cassandra Datasource Demo` dashboard. You are done! Feel free to play with the setup!

<img width="661" alt="image" src="https://user-images.githubusercontent.com/1742301/161745038-9221c48a-ed09-49c9-b7a6-eee73fa2a9ac.png">

## Play with the Datasource 

### 1. Check the datasource configuration

Open the link [http://localhost:3000/datasources](http://localhost:3000/datasources), and select the **Apache Cassandra** datasource. It's configured to use docker container but you can reconfigure it to use your own Cassandra Deployment if available.

<img width="425" alt="image" src="https://user-images.githubusercontent.com/1742301/161746033-17b05bdc-0687-4ac4-b919-9da7d773737f.png">

### 2. Review the Query Configurator panel

Query Configurator is the simplest way to configure data visualization, just paste the column names for the datasource to use and enjoy.

<img width="960" alt="image" src="https://user-images.githubusercontent.com/1742301/161747540-6490a24c-091b-40a0-8981-15cf6a92a9a8.png">

### 3. Review the Query Editor panel

Query Editor is a more complex but also more powerful way to work with your data, with the Editor you can use all `CQL` features including UDTs.

<img width="956" alt="image" src="https://user-images.githubusercontent.com/1742301/161747465-670068ef-1144-4d66-a0a4-c637a78188c0.png">

**That's it, you made it!**

## Don't forget to clean up!

Remove the running containers as soon as you don't need them.

```
docker-compose kill
docker-compose down
```

## Got questions? 

Contact us using [Github Discussions](https://github.com/HadesArchitect/GrafanaCassandraDatasource/discussions)!

-------
Lead maintainer [Aleks Volochnev](https://www.linkedin.com/in/aleks-volochnev/)