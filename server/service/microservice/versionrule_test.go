//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package microservice

import (
        . "github.com/onsi/ginkgo"
        . "github.com/onsi/gomega"
        "sort"
        "github.com/coreos/etcd/mvcc/mvccpb"
        "fmt"
)

var _ = Describe("Version Rule sorter", func() {
        Describe("Sorter", func() {
                Context("Normal", func() {
                        It("version asc", func() {
                                kvs := []string{"1.0.0", "1.0.1"}
                                sort.Sort(&serviceKeySorter{
                                        sortArr: kvs,
                                        kvs:     make(map[string]string),
                                        cmp:     Larger,
                                })
                                Expect(kvs[0]).To(Equal("1.0.1"))
                                Expect(kvs[1]).To(Equal("1.0.0"))
                        })
                        It("version desc", func() {
                                kvs := []string{"1.0.1", "1.0.0"}
                                sort.Sort(&serviceKeySorter{
                                        sortArr: kvs,
                                        kvs:     make(map[string]string),
                                        cmp:     Larger,
                                })
                                Expect(kvs[0]).To(Equal("1.0.1"))
                                Expect(kvs[1]).To(Equal("1.0.0"))
                        })
                        It("len(v1) != len(v2)", func() {
                                kvs := []string{"1.0.0.0", "1.0.1"}
                                sort.Sort(&serviceKeySorter{
                                        sortArr: kvs,
                                        kvs:     make(map[string]string),
                                        cmp:     Larger,
                                })
                                Expect(kvs[0]).To(Equal("1.0.1"))
                                Expect(kvs[1]).To(Equal("1.0.0.0"))
                        })
                        It("1.0.9 vs 1.0.10", func() {
                                kvs := []string{"1.0.9", "1.0.10"}
                                sort.Sort(&serviceKeySorter{
                                        sortArr: kvs,
                                        kvs:     make(map[string]string),
                                        cmp:     Larger,
                                })
                                Expect(kvs[0]).To(Equal("1.0.10"))
                                Expect(kvs[1]).To(Equal("1.0.9"))
                        })
                        It("1.10 vs 4", func() {
                                kvs := []string{"1.10", "4"}
                                sort.Sort(&serviceKeySorter{
                                        sortArr: kvs,
                                        kvs:     make(map[string]string),
                                        cmp:     Larger,
                                })
                                Expect(kvs[0]).To(Equal("4"))
                                Expect(kvs[1]).To(Equal("1.10"))
                        })
                })
                Context("Exception", func() {
                        It("invalid version1", func() {
                                kvs := []string{"1.a", "1.0.1.a"}
                                sort.Sort(&serviceKeySorter{
                                        sortArr: kvs,
                                        kvs:     make(map[string]string),
                                        cmp:     Larger,
                                })
                                Expect(kvs[0]).To(Equal("1.a"))
                                Expect(kvs[1]).To(Equal("1.0.1.a"))
                        })
                        It("invalid version2 > 127", func() {
                                kvs := []string{"1.0", "1.0.1.128"}
                                sort.Sort(&serviceKeySorter{
                                        sortArr: kvs,
                                        kvs:     make(map[string]string),
                                        cmp:     Larger,
                                })
                                Expect(kvs[0]).To(Equal("1.0"))
                                Expect(kvs[1]).To(Equal("1.0.1.128"))
                        })
                })
        })
        Describe("VersionRule", func() {
                Context("Normal", func() {
                        const count = 10
                        var kvs = [count]*mvccpb.KeyValue{}
                        BeforeEach(func() {
                                for i := 1; i<= count; i++ {
                                        kvs[i-1] = &mvccpb.KeyValue{
                                                Key: []byte(fmt.Sprintf("/service/ver/1.%d", i)),
                                                Value: []byte(fmt.Sprintf("%d", i)),
                                        }
                                }
                        })
                        It("Latest", func() {
                                results := VersionRule(Latest).GetServicesIds(kvs[:])
                                Expect(len(results)).To(Equal(1))
                                Expect(results[0]).To(Equal(fmt.Sprintf("%d", count)))
                        })
                        It("Range1.1 ver in [1.4, 1.8]", func() {
                                results := VersionRule(Range).GetServicesIds(kvs[:], "1.4", "1.8")
                                Expect(len(results)).To(Equal(5))
                                Expect(results[0]).To(Equal("8"))
                                Expect(results[4]).To(Equal("4"))
                        })
                        It("Range1.2 ver in [1.8, 1.4]", func() {
                                results := VersionRule(Range).GetServicesIds(kvs[:], "1.8", "1.4")
                                Expect(len(results)).To(Equal(5))
                                Expect(results[0]).To(Equal("8"))
                                Expect(results[4]).To(Equal("4"))
                        })
                        It("Range2 ver in [1, 2]", func() {
                                results := VersionRule(Range).GetServicesIds(kvs[:], "1", "2")
                                Expect(len(results)).To(Equal(10))
                                Expect(results[0]).To(Equal("10"))
                                Expect(results[9]).To(Equal("1"))
                        })
                        It("Range3 ver in [1.4.1, 1.9.1]", func() {
                                results := VersionRule(Range).GetServicesIds(kvs[:], "1.4.1", "1.9.1")
                                Expect(len(results)).To(Equal(5))
                                Expect(results[0]).To(Equal("9"))
                                Expect(results[4]).To(Equal("5"))
                        })
                        It("Range4 ver in [2, 4]", func() {
                                results := VersionRule(Range).GetServicesIds(kvs[:], "2", "4")
                                Expect(len(results)).To(Equal(0))
                        })
                        It("AtLess1 ver >= 1.6", func() {
                                results := VersionRule(AtLess).GetServicesIds(kvs[:], "1.6")
                                Expect(len(results)).To(Equal(5))
                                Expect(results[0]).To(Equal("10"))
                                Expect(results[4]).To(Equal("6"))
                        })
                        It("AtLess2 ver >= 1", func() {
                                results := VersionRule(AtLess).GetServicesIds(kvs[:], "1")
                                Expect(len(results)).To(Equal(10))
                                Expect(results[0]).To(Equal("10"))
                                Expect(results[9]).To(Equal("1"))
                        })
                        It("AtLess3 ver >= 1.5.1", func() {
                                results := VersionRule(AtLess).GetServicesIds(kvs[:], "1.5.1")
                                Expect(len(results)).To(Equal(5))
                                Expect(results[0]).To(Equal("10"))
                                Expect(results[4]).To(Equal("6"))
                        })
                        It("AtLess4 ver >= 2", func() {
                                results := VersionRule(AtLess).GetServicesIds(kvs[:], "2")
                                Expect(len(results)).To(Equal(0))
                        })
                })
                Context("Exception", func() {
                        It("nil", func() {
                                results := VersionRule(Latest).GetServicesIds(nil)
                                Expect(len(results)).To(Equal(0))
                                results = VersionRule(AtLess).GetServicesIds(nil)
                                Expect(len(results)).To(Equal(0))
                                results = VersionRule(Range).GetServicesIds(nil)
                                Expect(len(results)).To(Equal(0))
                        })
                })
        })
})
