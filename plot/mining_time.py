import numpy as np
import matplotlib.pyplot as plt

data = np.loadtxt("../stats/mining_time_16")
data = data / 1000000000
mean = np.mean(data)
var = np.var(data)
s_mean = str(mean)
s_var = str(var)
plt.figure(1)
plt.hist(data, label="Mean = " + s_mean[:5] + "s, Var = " + s_var[:5] + "s")
plt.xlabel("Mining time (s)")
plt.ylabel("Number of events")
plt.title("Histogram of mining time. TBITS = 16.")
plt.legend()
plt.savefig("mining_time_16.png")