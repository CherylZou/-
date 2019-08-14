extern {
    fn getValuefromNode1(x: i32) -> i32;
}

extern {
    fn getValuefromNode2(x: i32) -> i32;
}

extern {
    fn getValuefromNode(nodeid: i32 , x: i32) -> i32;
}

#[no_mangle]
pub extern fn addsum(x: i32, y: i32) -> i32 {
    unsafe { getValuefromNode1(x) + getValuefromNode2(y) }
}

#[no_mangle]
pub extern fn addsum2(nodeid1: i32, x1: i32, nodeid2: i32, x2: i32) -> i32 {
    unsafe { getValuefromNode(nodeid1,x1) + getValuefromNode(nodeid2,x2) }
}

