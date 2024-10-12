identifier=$1

# Check if /osdata/osgroup10 exists
rm -rf /osdata/osgroup10
mkdir -p /osdata/osgroup10

cd ~/Project2
stdbuf -oL ./main -mid=$identifier > "debug${identifier}.out" 2>"error${identifier}.out"