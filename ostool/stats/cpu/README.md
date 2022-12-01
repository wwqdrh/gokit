> 单位，jiffies是内核中的一个全局变量，用来记录自系统启动一来产生的节拍数，在linux中，一个节拍大致可理解为操作系统进程调度的最小时间片，不同linux内核可能值有不同，通常在1ms到10ms之间

# /proc/cpuinfo

系统中CPU的提供商和相关配置信息

processor　：系统中逻辑处理核心数的编号，从0开始排序。

vendor_id　：CPU制造商

cpu family　：CPU产品系列代号

model　　　：CPU属于其系列中的哪一代的代号

model name：CPU属于的名字及其编号、标称主频

stepping　 ：CPU属于制作更新版本

cpu MHz　 ：CPU的实际使用主频

cache size ：CPU二级缓存大小

physical id ：单个物理CPU的标号

siblings ：单个物理CPU的逻辑CPU数。siblings=cpu cores [*2]。

core id ：当前物理核在其所处CPU中的编号，这个编号不一定连续。

cpu cores ：该逻辑核所处CPU的物理核数。比如此处cpu cores 是4个，那么对应core id 可能是 1、3、4、5。

apicid ：用来区分不同逻辑核的编号，系统中每个逻辑核的此编号必然不同，此编号不一定连续

fpu ：是否具有浮点运算单元（Floating Point Unit）

fpu_exception ：是否支持浮点计算异常

cpuid level ：执行cpuid指令前，eax寄存器中的值，根据不同的值cpuid指令会返回不同的内容

wp ：表明当前CPU是否在内核态支持对用户空间的写保护（Write Protection）

flags ：当前CPU支持的功能

bogomips：在系统内核启动时粗略测算的CPU速度（Million Instructions Per Second

clflush size ：每次刷新缓存的大小单位

cache_alignment ：缓存地址对齐单位

address sizes ：可访问地址空间位数

power management ：对能源管理的支持

# /proc/stat

这个文件包含了所有CPU活动的信息

```bash
cpu user nice system idle iowait irq softirq
cpu0
cpu1
...
cpu4
...
intr 给出中断的信息，第一个为自系统启动以来，发生的所有的中断的次数；然后每个数对应一个特定的中断自系统启动以来所发生的次数。
ctxt 自系统启动以来CPU发生的上下文交换的次数。
btime 从系统启动到现在为止的时间，单位为秒
processes 自系统启动以来所创建的任务的个数目。
procs_running 当前运行队列的任务的数目。
procs_blocked 当前被阻塞的任务的数目。
softirq 从系统启动开始累计到当前时刻，软中断时间
```

