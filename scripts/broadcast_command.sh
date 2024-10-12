START_PORT=8060
USERNAME="osgroup10"
num_servers=$1

if [ -z "$num_servers" ]; then
    echo "did not specify number of servers, use 1 by default"
    num_servers=1
fi

command=${*:2}

if [ -z "$command" ]; then
    echo "did not specify command, use 'ls' by default"
    command="ls"
fi

# Generate SSH addresses and run the command
for (( i=START_PORT; i<START_PORT+num_servers; i++ )); do
    ssh -o StrictHostKeyChecking=No "${USERNAME}@122.200.68.26" -p $i "$command" &
done

# Wait for all background processes to finish
wait
