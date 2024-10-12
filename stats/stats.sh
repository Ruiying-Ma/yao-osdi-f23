#!/bin/bash

# Extract Mining time from debugxxxx.out
echo > mining_time_16
grep "Mining time" ../debug8060.out | awk '{print $4}' >> mining_time_16
grep "Mining time" ../debug8061.out | awk '{print $4}' >> mining_time_16
grep "Mining time" ../debug8062.out | awk '{print $4}' >> mining_time_16
grep "Mining time" ../debug8063.out | awk '{print $4}' >> mining_time_16
grep "Mining time" ../debug8064.out | awk '{print $4}' >> mining_time_16

# Extract Verifying block time from debugxxxx.out
echo > verifying_block_time
grep "Verifying block time" ../debug8060.out | awk '{print $5}' >> verifying_block_time
grep "Verifying block time" ../debug8061.out | awk '{print $5}' >> verifying_block_time
grep "Verifying block time" ../debug8062.out | awk '{print $5}' >> verifying_block_time
grep "Verifying block time" ../debug8063.out | awk '{print $5}' >> verifying_block_time
grep "Verifying block time" ../debug8064.out | awk '{print $5}' >> verifying_block_time
