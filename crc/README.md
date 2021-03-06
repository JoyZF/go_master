# CRC
## CRC是什么？
我们进行通信时的网络信道并不总是可靠的。为了增加可靠性，我们需要在传输数据后加上一些冗余的码字。如果接收方能够通过它们直接纠正错误，那么我们就称之为纠错码（Error Correcting Code）。如果接收方仅能通过它们发现错误，而真正纠正错误的过程需要通知发送方进行重传，那么我们就称之为检错码（Error Detecting Code）。纠错码的代价比检错码高，这一点只要看一下两者对汉明距离的要求就可以了：


当最小汉明距离大于0时，能够纠正的错误数一定小于能够发现的错误数。

其实凭直觉也能想到：纠正错误所需要的信息量应该是大于发现错误需要的信息量的。

用哪一种码字取决于信道的可靠程度（每个比特的错误概率）。当信道非常可靠的时候，检错码代价就会小于纠错码。（检错码较短，虽然一旦错误了需要重传，但重传的概率非常低，所以总体花费较少）相关的阈值计算在Computer Networks, 5 Edition 第三章课后习题有，不再赘述。

我们设计冗余码的原则就是：用最小的代价，检测（纠正）最多的错误。


CRC就是一种优秀的检错码。它的计算原理，说白了就是作除法。把比特流看作多项式的系数。设定一个生成多项式（generator polynomial）作为除数。数据流看作被除数。发送方需要在数据流末尾加上一段冗余码，使得组合后的新数据流能够整除除数。这段冗余码就是所谓的CRC（如何计算？在数据流末尾补CRC长度的0，然后做除法得到的余数就是了。）发送方计算好CRC后，把它加到末尾。然后接收方通过传过来的数据做除法计算余数，如果余数不为0，就说明有错误发生。