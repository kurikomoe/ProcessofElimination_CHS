#include <cstdint>
#include <cstddef>

using u8  = uint8_t;
using u16 = uint16_t;
using u32 = uint32_t;
using u64 = uint64_t;

using i8  = int8_t;
using i16 = int16_t;
using i32 = int32_t;
using i64 = int64_t;


struct TalkEntryData {
    u8* raw;
};

struct TalkEntry {
    u32  id;
    u32  dwUn0;
    u32  offset;
};

struct TalkDat {
    u32  dwEntrySize;
    u32  dwEntryCount;

    TalkEntry*      aEntries;
    TalkEntryData*  aEntryData;
};
