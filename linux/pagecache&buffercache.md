# Linux系统中的Page cache和Buffer cache

## Free命令显示内存
![](_docs/8153732@2x.png)

首先，我们来了解下内存的使用情况

![](_docs/img.png)

- Mem：表示物理内存统计。
- total：表示物理内存总量(total = used + free)。
- used：表示总计分配给缓存（包含buffers 与cache ）使用的数量，但其中可能部分缓存并未实际使用。
- free：未被分配的内存
- shared：共享内存
- buffers：系统分配但未被使用的buffers数量。
- cached：系统分配但未被使用的cache数量。
- -/+ buffers/cache：表示物理内存的缓存统计。
- used2：也就是第一行中的used – buffers - cached也是实际使用的内存总量。 // used2为第二行
- free2 = buffers1 + cached1 + free1 // free2为第二行，buffers1等为第一行
- free2：未被使用的buffers与cache和未被分配的内存之和，这就是系统当前实际可用内存。
- Swap：表示硬盘上交换分区的使用情况。

在Free命令中显示的buffer和cache，它们都是占用内存：

buffer : 作为buffer cache的内存，是块设备的读写缓冲区，更靠近存储设备，或者直接就是disk的缓冲区。

cache: 作为page cache的内存, 文件系统的cache，是memory的缓冲区 。

如果cache 的值很大，说明cache住的文件数很多。如果频繁访问到的文件都能被cache住，那么磁盘的读IO 必会非常小 。

## Page cache（页面缓存）

Page cache 也叫页缓冲或文件缓冲，是由好几个磁盘块构成，大小通常为4k，在64位系统上为8k，构成的几个磁盘块在物理磁盘上不一定连续，文件的组织单位为一页， 也就是一个page cache大小，文件读取是由外存上不连续的几个磁盘块，到buffer cache，然后组成page cache，然后供给应用程序。

Page cache在linux读写文件时，它用于缓存文件的逻辑内容，从而加快对磁盘上映像和数据的访问。具体说是加速对文件内容的访问，buffer cache缓存文件的具体内容——物理磁盘上的磁盘块，这是加速对磁盘的访问。

## Buffer cache（块缓存）

Buffer cache 也叫块缓冲，是对物理磁盘上的一个磁盘块进行的缓冲，其大小为通常为1k，磁盘块也是磁盘的组织单位。设立buffer cache的目的是为在程序多次访问同一磁盘块时，减少访问时间。系统将磁盘块首先读入buffer cache，如果cache空间不够时，会通过一定的策略将一些过时或多次未被访问的buffer cache清空。程序在下一次访问磁盘时首先查看是否在buffer cache找到所需块，命中可减少访问磁盘时间。不命中时需重新读入buffer cache。对buffer cache的写分为两种，一是直接写，这是程序在写buffer cache后也写磁盘，要读时从buffer cache上读，二是后台写，程序在写完buffer cache后并不立即写磁盘，因为有可能程序在很短时间内又需要写文件，如果直接写，就需多次写磁盘了。这样效率很低，而是过一段时间后由后台写，减少了多次访磁盘的时间。

Buffer cache是由物理内存分配，Linux系统为提高内存使用率，会将空闲内存全分给buffer cache ，当其他程序需要更多内存时，系统会减少cache大小。


## Buffer page（缓冲页）

如果内核需要单独访问一个块，就会涉及到buffer page，并会检查对应的buffer head。

## Swap space（交换空间）

Swap space交换空间，是虚拟内存的表现形式。系统为了应付一些需要大量内存的应用，而将磁盘上的空间做内存使用，当物理内存不够用时，将其中一些暂时不需的数据交换到交换空间，也叫交换文件或页面文件中。做虚拟内存的好处是让进程以为好像可以访问整个系统物理内存。因为在一个进程访问数据时，其他进程的数据会被交换到交换空间中。

## Swap cache（交换缓存）

swap cache，它表示交换缓存的大小。Page cache是磁盘数据在内存中的缓存，而swap cache则是交换分区在内存中的临时缓存。

## Memory mapping（内存映射）

内核有两种类型的内存映射：共享型(shared)和私有型(private)。私有型是当进程为了只读文件，而不写文件时使用，这时，私有映射更加高效。 但是，任何对私有映射页的写操作都会导致内核停止映射该文件中的页。所以，写操作既不会改变磁盘上的文件，对访问该文件的其它进程也是不可见的。

共享内存中的页通常都位于page cache，私有内存映射只要没有修改，也位于page cache。当进程试图修改一个私有映射内存页时，内核就把该页进行复制，并在页表中用复制的页替换原来的页。由于修改了页表，尽管原来的页仍然在 page cache，但是已经不再属于该内存映射。而新复制的页也不会插入page cache，而是添加到匿名页反向映射数据结构。

## Page cache和Buffer cache的区别

磁盘的操作有逻辑级（文件系统）和物理级（磁盘块），这两种Cache就是分别缓存逻辑和物理级数据的。

也就是说，直接去操作文件，那就是Page Cache区缓存，用dd等命令直接操作磁盘块，就是Buffer Cache缓存的东西。



Page cache实际上是针对文件系统的，是文件的缓存，在文件层面上的数据会缓存到page cache。文件的逻辑层需要映射到实际的物理磁盘，这种映射关系由文件系统来完成。当page cache的数据需要刷新时，page cache中的数据交给buffer cache，但是这种处理在2.6版本的内核之后就变的很简单了，没有真正意义上的cache操作。



Buffer cache是针对磁盘块的缓存，也就是在没有文件系统的情况下，直接对磁盘进行操作的数据会缓存到buffer cache中，例如，文件系统的元数据都会缓存到buffer cache中。



简单说来，page cache用来缓存文件数据，buffer cache用来缓存磁盘数据。在有文件系统的情况下，对文件操作，那么数据会缓存到page cache，如果直接采用dd等工具对磁盘进行读写，那么数据会缓存到buffer cache。



Buffer(Buffer Cache)以块形式缓冲了块设备的操作，定时或手动的同步到硬盘，它是为了缓冲写操作然后一次性将很多改动写入硬盘，避免频繁写硬盘，提高写入效率。

