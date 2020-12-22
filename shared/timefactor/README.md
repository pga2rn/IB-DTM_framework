# Time factor

time factor is for tuning the raw value of trust value offset when generating trust value.

## Trust value calculation

RSUs query all trust value offsets of specific vehicles within an epoch.

Trust value $V_i$ of vehicle $i$ in epoch $E$  can be calculated as follow, $S$ is the number of slots in $E$, $v_s$ is the sum of  trust value offsets for slot $s$, function $f$ is the time factor function, taking the start time $t_s$ of slot $s$ as parameter. 
$$
V_i=\sum^{S}_{s=0}{v_s \times f(t_s)},\quad V_i \in [-1, 1]
$$

## time factor function

There are 5 functions available in the simulation

- Exponential function: $ y = 2^x - 1$
- Linear function: $ y = x$
- Power function: $y = x^2$
- Sin function: $y = sin(\frac{pi \times x}{2})$
- Log function: $y = -\frac{1}{2} \times log(\frac{1}{x}) + 1$

These function take $x$ from range $[0, 1]$ and yield results $y$ range from $[0, 1]$( negative $y$ will be rounded up to 0). 

Where $x$ can be calculated from $T_{genesis}$, $t_{nextEpoch}$, and $t_{slot}$ as follow:
$$
x = (t_{slot} - T_{genesis})/(t_{nextEpoch} - T_{genesis})
$$
$x$ represents the position of slot $s$ from $genesis$.

The idea is that the newer the trust value offsets the more it will reflect on the final trust value result.