package main

import "fmt"

// 实现删除切片特定下标元素的方法。
// 输入参数：
// src：  待删除元素的切片
// idx：  删除元素的下标
// 输出参数：
// 参数1： 删除后的切片
// 参数2： 删除的元素
// 参数3： 错误信息

// DeleteIdxV1 能够实现删除操作就可以
func DeleteIdxV1(src []int, idx int) ([]int, int, error) {
	if idx < 0 || idx >= len(src) {
		return src, 0, fmt.Errorf("删除下标错误")
	}

	res := src[idx]

	j := 0
	for i, v := range src {
		if i != idx {
			src[j] = v
			j++
		}
	}
	src = src[:j]

	return src, res, nil
}

// DeleteIdxV2 考虑使用比较高性能的实现
func DeleteIdxV2(src []int, idx int) ([]int, int, error) {
	if idx < 0 || idx >= len(src) {
		return src, 0, fmt.Errorf("删除下标错误")
	}

	res := src[idx]

	return append(src[:idx], src[idx+1:]...), res, nil
}

// DeleteIdxV3 改造为泛型方法
func DeleteIdxV3[T any](src []T, idx int) ([]T, T, error) {
	var res T
	if idx < 0 || idx >= len(src) {
		return src, res, fmt.Errorf("删除下标错误")
	}

	res = src[idx]

	return append(src[:idx], src[idx+1:]...), res, nil
}

// DeleteIdxV4 支持缩容，并且设计缩容机制
// 我的缩容算法，当切片长度小于切片容量的1/3时触发切片的缩容，新切片的容量为老切片的一半
func DeleteIdxV4[T any](src []T, idx int) ([]T, T, error) {
	var res T
	if idx < 0 || idx >= len(src) {
		return src, res, fmt.Errorf("删除下标错误")
	}

	res = src[idx]

	//比较切片的长度和容量，以确定是否需要触发缩容
	if len(src) <= cap(src)/3 {
		//缩容
		fmt.Printf("开始缩容，旧切片的容量为：%d，长度为：%d\n", cap(src), len(src))

		dest := make([]T, len(src)-1, cap(src)/2)
		j := 0
		for i := 0; i < len(src); i++ {
			if i != idx {
				dest[j] = src[i]
				j++
			}
		}

		fmt.Printf("缩容结束，新切片的容量为：%d，长度为：%d\n", cap(dest), len(dest))

		return dest, res, nil
	} else {
		//不触发缩容，则返回原切片
		return append(src[:idx], src[idx+1:]...), res, nil
	}

}

func main() {

	var res int
	var resStr string

	sliceIntV1 := []int{1, 2, 3, 4, 5, 6, 7}
	var errV1 error
	var idxV1 int = 3

	sliceIntV1, res, errV1 = DeleteIdxV1(sliceIntV1, idxV1)
	if errV1 != nil {
		fmt.Println(errV1)
	} else {
		fmt.Printf("DeleteIdxV1: 删除后的切片为： %v; 删除的下标是：%d, 删除的元素是： %d\n", sliceIntV1, idxV1, res)
	}

	sliceIntV2 := []int{1, 2, 3, 4, 5, 6, 7, 8}
	var errV2 error
	var idxV2 int = 5

	sliceIntV2, res, errV2 = DeleteIdxV2(sliceIntV2, idxV2)
	if errV2 != nil {
		fmt.Println(errV2)
	} else {
		fmt.Printf("DeleteIdxV2: 删除后的切片为： %v; 删除的下标是：%d, 删除的元素是： %d\n", sliceIntV2, idxV2, res)
	}

	sliceStringV3 := []string{"one", "two", "three", "four", "five", "six", "seven"}
	var errV3 error
	var idxV3 int = 4
	sliceStringV3, resStr, errV3 = DeleteIdxV3(sliceStringV3, idxV3)
	if errV3 != nil {
		fmt.Println(errV3)
	} else {
		fmt.Printf("DeleteIdxV3: 删除后的切片为： %v; 删除的下标是：%d, 删除的元素是： %s\n", sliceStringV3, idxV3, resStr)
	}

	sliceStringV4 := make([]string, 2, 7)
	sliceStringV4[0] = "one"
	sliceStringV4[1] = "two"
	var errV4 error
	var idxV4 int = 0

	sliceStringV4, resStr, errV4 = DeleteIdxV4(sliceStringV4, idxV4)
	if errV4 != nil {
		fmt.Println(errV4)
	} else {
		fmt.Printf("DeleteIdxV4: 删除后的切片为： %v; 删除的下标是：%d, 删除的元素是： %s\n", sliceStringV4, idxV4, resStr)
	}
}
