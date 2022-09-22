Storj placement simulator

This simple project can help to evaluate different placement policy on the storage network.

Requirements:

1. You need the export of the node table (`node.csv`) from a satellite
2. You should execute the `placement_test.go` unit test
3. results are printed out / saved to csv files.

## Example output

The raw output from an execution of 1 000 000 placement is copied to here. To understand the numbers better, I copied a
more detailed mail to the next section where I explained these in more details.

```
-----------State------------
node influence
max: 1
1 80000000

net influence
max: 1
1 80000000

wallet influence
max: 13
11 60
12 9
1 66751120
4 416116
7 17462
8 4904
9 1198
10 271
3 1099202
2 3501358
5 156637
6 54195
13 1

State by node
13782
StdDev: 3533.748766
Min: 8.000000
Max: 9466.000000
Mean: 5804.672762

State by wallet
5439
StdDev: 66899.940507
Min: 29.000000
Max: 2676514.000000
Mean: 14708.586137

State by net
9723
StdDev: 2437.520594
Min: 2108.000000
Max: 11606.000000
Mean: 8227.913196

-----------RandomSelector------------
node influence
max: 3
1 79486611
2 255820
3 583

net influence
max: 6
1 78387598
2 763503
3 26224
4 1558
5 90
6 7

wallet influence
max: 12
2 3946179
5 74586
6 19583
10 37
9 174
11 3
1 67775954
3 913606
8 919
7 4771
12 1
4 264428

RandomSelector by node
13782
StdDev: 2090.984918
Min: 1375.000000
Max: 7124.000000
Mean: 5804.672762

RandomSelector by wallet
5439
StdDev: 57529.804577
Min: 1375.000000
Max: 2035969.000000
Mean: 14708.586137

RandomSelector by net
9723
StdDev: 10589.115256
Min: 1375.000000
Max: 409805.000000
Mean: 8227.913196

-----------SpaceSelector------------
node influence
max: 4
1 79037258
2 477850
3 2338
4 7

net influence
max: 4
2 571055
3 3419
4 16
1 78847569

wallet influence
max: 12
1 66383677
2 4047334
6 33727
7 9275
8 2189
10 95
11 22
3 1082668
4 354802
5 112719
9 537
12 2

SpaceSelector by node
13678
StdDev: 6136.717110
Min: 1.000000
Max: 23909.000000
Mean: 5848.808305

SpaceSelector by wallet
5398
StdDev: 63537.510222
Min: 1.000000
Max: 2350039.000000
Mean: 14820.303816

SpaceSelector by net
9667
StdDev: 7315.483840
Min: 1.000000
Max: 82411.000000
Mean: 8275.576704
```

## Explanation

I think a good placement policy:

* (addressing node loosing risk) should distribute the pieces of each segment to as much node as possible
* (addressing network loosing risk) should distribute the pieces of each segment to as much network as possible
* (addressing owner loosing risk) should distribute the pieces of each segment to as much wallet as possible
* (being fair) Node owner with huge number of nodes shouldn't have proportional higher chance to get data.

* I think (1) and (3) risks are higher/more important than 2, but it's subjective opinion.

* But all of these can be defined mathematically and proven with simulation.

* For example let's imagine that I do node selection of 80 nodes 1 000 000 times.
  (used node table from a few months ago: 11140 reputable and 2642 new nodes)

With the current algorithm

* Got 1 000 000 segments, and none of them have two pieces on the same net
* It occured 1 times that 13 pieces (out of 80 in one segment) are assigned to nodes with the same wallet. (9 times for
  12 pieces, ...)
* One node will get 5.8k  (mean) pieces with std dev 3.5k
* One wallet will get 14.7k (mean) pieces with std dev 67k
* One network will get 8.2K pieces with std dev 2.4k

Let's replace the current selection with totally random selection (ignoring network constraint, even ignoring any same
node constraint)

* 7 times I will have such placements where 6 selected nodes share the same network. (90 times for 5).
* once 11 pieces are assigned to nodes with the same wallet
* 583 times we put 3 pieces to the same node
* One node will get 5.8k  (mean) pieces with std dev 2 k
* One wallet will get 14.7 k (mean) pieces with std dev 57k
* One network will get 8.2K pieces with std dev 10k

## Summary:

* node loosing risk: random algorithm is slightly worse, but worst case we used the same node for 3 pieces, which is
  still very safe because our generous EC numbers.
* net loosing risk: same is true here. Worst case we used 6 nodes from the same net which can be survived. And typical
  number is 2-4, but most of the time is 1
* wallet loosing risk: surprise, the random selection is slightly better. We distributed data better between wallets.
* fairness: the random algorithm is more fair with nodes and wallets. While the mean count of pieces are the same for
  wallets/nodes, the std dev is lower for the random selection: it means that we have less "unlucky" nodes/node groups
  which can get less portion from the new data.
* From this point of view, the fully random algorithm seems to be acceptable for me. It's more risky only with a very
  small percentage, which is fine with the current EC numbers. Same time it's slightly more fair and distributes
  slightly better the data between wallet owners.

But random selection also has two additional property:

* It removes incentives to create proxy nodes --> real node distribution will be more transparent
* It can be attacked easily with starting 100x processes on the same node (but we can improve it with node uniqueness)