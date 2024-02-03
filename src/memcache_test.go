package memcache

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const defaultAddr = "127.0.0.1:11211"

var _ = Describe("Memcache Client Tests", Label("StorageCommands"), func() {
	var mc *Client
	var it1 *Item
	var invalidIt *Item
	var appendIt *Item
	var prependIt *Item
	var lowExpIt *Item
	var toDeleteIt *Item
	var toIncr *Item

	BeforeEach(func() {
		mc = New([]string{defaultAddr})

		it1 = &Item{
			Key:        "hello",
			Value:      []byte("world"),
			Expiration: time.Second * 60,
			Flags:      0,
			CAS:        0,
		}
		invalidIt = &Item{
			Key:        "invalid key", // due to space
			Value:      []byte("world"),
			Expiration: time.Second * 60,
			Flags:      0,
			CAS:        0,
		}
		appendIt = &Item{
			Key:        "hello",
			Value:      []byte(" from memcache-go"),
			Expiration: time.Second * 60,
			Flags:      0,
			CAS:        0,
		}
		prependIt = &Item{
			Key:        "hello",
			Value:      []byte("Hello "),
			Expiration: time.Second * 60,
			Flags:      0,
			CAS:        0,
		}
		lowExpIt = &Item{
			Key:        "low_exp",
			Value:      []byte("quickly!"),
			Expiration: time.Second * 1,
			Flags:      0,
			CAS:        0,
		}
		toDeleteIt = &Item{
			Key:        "delete_me",
			Value:      []byte("nooow!"),
			Expiration: time.Second * 60,
			Flags:      0,
			CAS:        0,
		}
		toIncr = &Item{
			Key:        "incr_val",
			Value:      []byte("10"),
			Expiration: time.Second * 60,
			Flags:      0,
			CAS:        0,
		}
	})

	It("Memcache Commands with a working client", func() {
		By("Append on a key that doesn't exist")
		err := mc.Append(appendIt)
		Expect(err).To(HaveOccurred())

		By("Prepend on a key that doesn't exist")
		err = mc.Prepend(prependIt)
		Expect(err).To(HaveOccurred())

		By("Get after a Set command returns the correct value")
		err = mc.Set(it1)
		Expect(err).ToNot(HaveOccurred())
		res, err := mc.Get(it1.Key)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Value).To(Equal(it1.Value))
		Expect(res.Flags).To(Equal(it1.Flags))

		By("Append on a key that exists")
		err = mc.Append(appendIt)
		Expect(err).ToNot(HaveOccurred())
		newIt, err := mc.Get(appendIt.Key)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(newIt.Value)).To(Equal(string(it1.Value) + string(appendIt.Value)))

		By("Set an item with invalid key")
		err = mc.Set(invalidIt)
		Expect(err).To(HaveOccurred())

		By("Get fails after a set due to key's expiration")
		err = mc.Set(lowExpIt)
		Expect(err).ToNot(HaveOccurred())
		By("Waiting for the key to be deleted")
		time.Sleep(time.Second * 5)
		it, err := mc.Get(lowExpIt.Key)
		Expect(it).To(BeNil())
		Expect(err).To(HaveOccurred())

		By("Creating a new item and deleting it")
		err = mc.Set(toDeleteIt)
		Expect(err).ToNot(HaveOccurred())
		resDel, err := mc.Get(toDeleteIt.Key)
		Expect(err).ToNot(HaveOccurred())
		Expect(resDel.Value).To(Equal(toDeleteIt.Value))
		Expect(resDel.Flags).To(Equal(toDeleteIt.Flags))
		err = mc.Delete(resDel.Key)
		Expect(err).ToNot(HaveOccurred())
		_, err = mc.Get(toDeleteIt.Key)
		Expect(err).To(HaveOccurred())

		By("Creating a new item and incrementing it")
		err = mc.Set(toIncr)
		Expect(err).ToNot(HaveOccurred())
		newVal, err := mc.Incr(toIncr.Key, 10)
		Expect(err).ToNot(HaveOccurred())
		Expect(newVal).To(Equal(uint64(20)))

		By("Trying to increment a key that does not exist")
		_, err = mc.Incr("non_existing_key", 100)
		Expect(err).To(HaveOccurred())

		By("Trying to increment a key whose value is not numeric")
		_, err = mc.Incr(it1.Key, 100)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("cannot increment or decrement non-numeric value\r\n"))
	})
})
