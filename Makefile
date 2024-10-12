experiment: main
	bash scripts/launch_experiment.sh 5

stop: 
	bash scripts/stop_miners.sh 5

main: main.go
	go build main.go