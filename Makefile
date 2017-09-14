

goexample: 
	go run examples/golang/example_server.go

locustlocal:
	locust -f docker/locust-tasks/locustfile.py --host=http://localhost:8080

dockerexample:
	docker run -it -p=8080 --name=exampleserver --network=locustnw goexample

dockerlocal:
	docker run -it -p=8089:8089 -e "TARGET_HOST=http://exampleserver:8080" --network=locustnw locust-tasks:latest
