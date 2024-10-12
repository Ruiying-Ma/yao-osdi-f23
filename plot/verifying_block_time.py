import numpy as np
import matplotlib.pyplot as plt

data = np.loadtxt("../stats/verifying_block_time")
data = data / 1000000
mean = np.mean(data)
var = np.var(data)
s_mean = str(mean)
s_var = str(var)
plt.figure(1)
plt.hist(data, label="Mean = " + s_mean[:5] + "ms, Var = " + s_var[:5] + "ms")
plt.xlabel("Verifying block time (ms)")
plt.ylabel("Number of events")
plt.title("Histogram of verifying block time.")
plt.legend()
plt.savefig("verifying_block_time.png")