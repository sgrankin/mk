proto.txt:
	echo $target

prereq.(\w+):R: proto.${stem1}
	echo $stem1 <(echo ${stem1})
    