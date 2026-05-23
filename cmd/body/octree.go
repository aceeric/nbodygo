package body

import "math"

const (
	// Theta is the Barnes-Hut opening angle threshold. Lower = more accurate,
	// more expensive. 0.5 is the conventional default; 0.0 degenerates to
	// exact O(n²) via tree traversal.
	Theta float64 = 0.3

	// Softening prevents numerical blow-up when two bodies pass very close.
	// Replace distSq with distSq + softening² in the force denominator.
	// Tune to your simulation's length scale.
	Softening float64 = 1e-4

	// leafCapacity is the number of bodies a node holds before subdividing.
	leafCapacity int = 1

	// initialPoolSize is the starting capacity of the node pool slab.
	// Sized generously to avoid re-allocation in typical runs; the pool
	// grows automatically if needed.
	initialPoolSize int = 16384
)

// octNode is a node in the octree. It is either:
//   - a leaf node:     children[0] == nil, body != nil (or empty: body == nil)
//   - an internal node: children[0] != nil, body == nil
//
// Each node tracks the total mass and center of mass of all bodies it contains,
// which is what Barnes-Hut uses to approximate distant force.
type octNode struct {
	// Axis-aligned bounding box: center + half-size
	cx, cy, cz float64
	half       float64

	// Aggregate mass properties — updated incrementally as bodies are inserted
	totalMass float64
	cmX       float64
	cmY       float64
	cmZ       float64

	// Leaf body (nil if empty or internal)
	body *Body

	// Children indexed by octant: index = (z<<2 | y<<1 | x)
	//   x bit: 0 = left  (-x), 1 = right (+x)
	//   y bit: 0 = bottom(-y), 1 = top   (+y)
	//   z bit: 0 = front (-z), 1 = back  (+z)
	children [8]*octNode
}

func (n *octNode) isLeaf() bool { return n.children[0] == nil }

// octantOf returns the child index (0-7) for a given position
func (n *octNode) octantOf(x, y, z float64) int {
	idx := 0
	if x >= n.cx {
		idx |= 1
	}
	if y >= n.cy {
		idx |= 2
	}
	if z >= n.cz {
		idx |= 4
	}
	return idx
}

// childCenter returns the center of the given octant
func (n *octNode) childCenter(octant int) (cx, cy, cz float64) {
	q := n.half / 2
	if octant&1 != 0 {
		cx = n.cx + q
	} else {
		cx = n.cx - q
	}
	if octant&2 != 0 {
		cy = n.cy + q
	} else {
		cy = n.cy - q
	}
	if octant&4 != 0 {
		cz = n.cz + q
	} else {
		cz = n.cz - q
	}
	return
}

// updateMass incrementally updates aggregate mass and center of mass
func (n *octNode) updateMass(mass, x, y, z float64) {
	newTotal := n.totalMass + mass
	n.cmX = (n.cmX*n.totalMass + x*mass) / newTotal
	n.cmY = (n.cmY*n.totalMass + y*mass) / newTotal
	n.cmZ = (n.cmZ*n.totalMass + z*mass) / newTotal
	n.totalMass = newTotal
}

// Octree is the top-level structure. It owns the node pool slab and the root
// node. The pool is retained between cycles and reused to eliminate per-cycle
// heap allocation pressure.
type Octree struct {
	root *octNode
	// pool is a pre-allocated slab of nodes. On each cycle the tree is rebuilt
	// by resetting 'used' to 0 and overwriting nodes in place — no allocations.
	pool []octNode
	used int
}

// NewOctree creates an Octree with a pre-allocated node pool. Call this once
// at simulation startup and retain the pointer. Each cycle, call
// octree.Build(bodies) to rebuild the tree in place.
func NewOctree() *Octree {
	return &Octree{
		pool: make([]octNode, initialPoolSize),
		used: 0,
	}
}

// alloc returns a pointer to the next free node in the pool, growing the
// pool if necessary. The node is zeroed before being returned.
func (t *Octree) alloc() *octNode {
	if t.used >= len(t.pool) {
		// Pool exhausted — grow by doubling
		t.pool = append(t.pool, make([]octNode, len(t.pool))...)
	}
	n := &t.pool[t.used]
	*n = octNode{} // zero the node (reuse from previous cycle)
	t.used++
	return n
}

// Build rebuilds the octree from the passed body array in place, reusing the
// existing node pool slab. Call this once per compute cycle, sequentially,
// before dispatching workers.
func (t *Octree) Build(bodies []*Body) {
	// Reset pool — all previous nodes are logically freed
	t.used = 0
	t.root = nil

	if len(bodies) == 0 {
		return
	}

	// Compute tight bounding box over all existing bodies
	minX, minY, minZ := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	maxX, maxY, maxZ := -math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	count := 0
	for _, b := range bodies {
		if !b.Exists || b.fragmenting {
			continue
		}
		if b.X < minX {
			minX = b.X
		}
		if b.Y < minY {
			minY = b.Y
		}
		if b.Z < minZ {
			minZ = b.Z
		}
		if b.X > maxX {
			maxX = b.X
		}
		if b.Y > maxY {
			maxY = b.Y
		}
		if b.Z > maxZ {
			maxZ = b.Z
		}
		count++
	}
	if count == 0 {
		return
	}

	// Square bounding box sized to the largest extent, plus a margin
	cx := (minX + maxX) / 2
	cy := (minY + maxY) / 2
	cz := (minZ + maxZ) / 2
	half := math.Max(maxX-minX, math.Max(maxY-minY, maxZ-minZ)) / 2 * 1.05

	t.root = t.alloc()
	t.root.cx = cx
	t.root.cy = cy
	t.root.cz = cz
	t.root.half = half

	for _, b := range bodies {
		if b.Exists && !b.fragmenting {
			t.insert(t.root, b)
		}
	}
}

// insert adds body b into the subtree rooted at n, subdividing as needed.
// Uses t.alloc() for new nodes rather than heap allocation.
func (t *Octree) insert(n *octNode, b *Body) {
	n.updateMass(b.Mass, b.X, b.Y, b.Z)

	if n.isLeaf() {
		if n.body == nil {
			// Empty leaf — store body here
			n.body = b
			return
		}
		// Occupied leaf — subdivide and re-insert both bodies
		existing := n.body
		n.body = nil
		t.subdivide(n)
		t.insertIntoChild(n, existing)
		t.insertIntoChild(n, b)
		return
	}

	t.insertIntoChild(n, b)
}

// subdivide creates the 8 children of n using the pool allocator
func (t *Octree) subdivide(n *octNode) {
	childHalf := n.half / 2
	for i := 0; i < 8; i++ {
		cx, cy, cz := n.childCenter(i)
		child := t.alloc()
		child.cx = cx
		child.cy = cy
		child.cz = cz
		child.half = childHalf
		n.children[i] = child
	}
}

// insertIntoChild routes body b to the correct child octant of n
func (t *Octree) insertIntoChild(n *octNode, b *Body) {
	octant := n.octantOf(b.X, b.Y, b.Z)
	t.insert(n.children[octant], b)
}

// CalcForce computes the gravitational force on body b using Barnes-Hut
// approximation for non-sun bodies, and direct O(n) summation for the sun.
//
// This is called by Body.Compute() in place of the IterateOnce inner loop.
// It is read-only on the tree — safe to call concurrently from multiple
// goroutines without any additional locking.
//
// Returns accumulated force components fx, fy, fz.
func (t *Octree) CalcForce(b *Body, bc *BodyCollection) (fx, fy, fz float64) {
	if t.root == nil {
		return
	}

	if b.IsSun {
		// Special case: the sun computes force directly from every body.
		// This is O(n) for one body — acceptable, and avoids the pathological
		// deep tree traversal that occurs when a massive central body tries to
		// approximate the force from a surrounding shell of bodies.
		bc.IterateOnce(func(other *Body) {
			if other == b || !other.Exists || other.fragmenting {
				return
			}
			dfx, dfy, dfz := directForce(b, other)
			fx += dfx
			fy += dfy
			fz += dfz
		})
		return
	}

	// All non-sun bodies use Barnes-Hut tree traversal
	accumForce(t.root, b, &fx, &fy, &fz)
	return
}

// accumForce recursively traverses the octree accumulating gravitational force
// on body b. This is the Barnes-Hut decision point — read-only, concurrency safe.
func accumForce(n *octNode, b *Body, fx, fy, fz *float64) {
	if n == nil || n.totalMass == 0 {
		return
	}

	dx := n.cmX - b.X
	dy := n.cmY - b.Y
	dz := n.cmZ - b.Z
	distSq := dx*dx + dy*dy + dz*dz + Softening*Softening
	dist := math.Sqrt(distSq)

	// Barnes-Hut criterion: if this is a leaf, or if the cell is far enough
	// away to be treated as a single point mass, apply force directly.
	// Criterion: (2 * half) / dist < Theta  i.e. the cell's angular size
	// is small enough that approximation error is acceptable.
	if n.isLeaf() || (2*n.half/dist) < Theta {
		// Skip self-interaction at leaf level
		if n.isLeaf() && n.body == b {
			return
		}
		force := G * b.Mass * n.totalMass / distSq
		*fx += force * dx / dist
		*fy += force * dy / dist
		*fz += force * dz / dist
		return
	}

	// Cell is too close to approximate — recurse into children
	for _, child := range n.children {
		if child != nil && child.totalMass > 0 {
			accumForce(child, b, fx, fy, fz)
		}
	}
}

// directForce computes the exact gravitational force on body a from body b,
// returning the force components. Used for sun-body interactions and as a
// fallback. Softening is applied for numerical stability.
func directForce(a, b *Body) (fx, fy, fz float64) {
	dx := b.X - a.X
	dy := b.Y - a.Y
	dz := b.Z - a.Z
	distSq := dx*dx + dy*dy + dz*dz + Softening*Softening
	dist := math.Sqrt(distSq)
	force := G * a.Mass * b.Mass / distSq
	fx = force * dx / dist
	fy = force * dy / dist
	fz = force * dz / dist
	return
}
