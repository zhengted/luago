## 

- 由于luago项目中的GC相关依赖Go的GC，所以单独出一篇文章介绍lua的GC算法

### 三色解释

- 白色：对象为待访问状态，表示对象还没有被GC标记过，所有对象创建的初始状态。如果GC扫描后仍为白色，则表示该对象没有被系统的任何一个对象引用，可以回收空间
- 灰色：待扫描状态，表示对象已经被GC访问过，但是该对象引用的其他对象还没有被访问到
- 黑色：已扫描状态，表示该对象已经被GC访问过，并且该对象引用的其他对象也被访问过了

### 三色算法伪代码流程

```
// 以Lua5.1为例
// 初始化阶段
遍历Root结点中引用的对象，从白色置为灰色，并且放入灰色结点中

// 标记阶段
当灰色链表不为空（还有未扫描的元素）：
	取出一个对象标记为黑色
	遍历这个对象关联的其他所有对象：
		如果是白色
		标记为灰色并加入灰色链表中
		
// 回收阶段
遍历所有对象：
	如果为白色：
		这些对象都是没有被引用的对象，逐个回收
	否则：
		重新加入对象链表等待下一轮的GC检查
```

#### 创建对象时

- 此处根据版本不同，GC算法的写法不太相同

```cpp
// Lua 5.3 (lgc.c line.204)
/*
** create a new collectable object (with given type and size) and link
** it to 'allgc' list.
*/
GCObject *luaC_newobj (lua_State *L, int tt, size_t sz) {
  global_State *g = G(L);
  GCObject *o = cast(GCObject *, luaM_newobject(L, novariant(tt), sz));
  o->marked = luaC_white(g);
  o->tt = tt;
  o->next = g->allgc;
  g->allgc = o;
  return o;
}

// Lua 5.1 (lgc.c line.685)
void luaC_link (lua_State *L, GCObject *o, lu_byte tt) {
  global_State *g = G(L);
  o->gch.next = g->rootgc;
  g->rootgc = o;
  o->gch.marked = luaC_white(g);
  o->gch.tt = tt;
}
```

#### 初始化阶段

- 遍历Root链表上的节点将将他们的颜色从白变灰

```cpp
// Lua 5.1 (lgc.c line.500)
/* mark root set */
static void markroot (lua_State *L) {
  global_State *g = G(L);
  g->gray = NULL;
  g->grayagain = NULL;
  g->weak = NULL;
  markobject(g, g->mainthread);
  /* make global table be traversed before main stack */
  markvalue(g, gt(g->mainthread));
  markvalue(g, registry(L));
  markmt(g);
  g->gcstate = GCSpropagate;
}

// Lua 5.3 (lgc.c line.334)
/*
** mark root set and reset all gray lists, to start a new collection
*/
static void restartcollection (global_State *g) {
  g->gray = g->grayagain = NULL;
  g->weak = g->allweak = g->ephemeron = NULL;
  markobject(g, g->mainthread);
  markvalue(g, &g->l_registry);
  markmt(g);
  markbeingfnz(g);  /* mark any finalizing object left from previous cycle */
}
```

- 针对不同的数据类型，置灰方式不同
  - 如字符串类型没有引用其他数据，可以跳过置灰阶段，将非黑的直接回收
  - userdata同理
  - upval则需要根据其open状态，如果非open可以直接标记为黑色，因为没有与其他数据的引用关系了

#### 扫描标记阶段

