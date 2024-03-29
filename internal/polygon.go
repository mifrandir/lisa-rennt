package internal

import (
	"errors"
	"strings"

	"github.com/miltfra/tools"
)

// A Polygon is a sequence of points (corners)
// connected to their respective neighbours.
type Polygon struct {
	Corners []*Corner
	N       int
}

// A Corner is a part of a polygon that has two
// neighbours and it's own poisition.
type Corner struct {
	*Point
	Plygn *Polygon
	N1    *Corner
	N2    *Corner
}

// NewPolygon returns an empty polygon.
func NewPolygon() *Polygon {
	p := &Polygon{make([]*Corner, 0, 3), 0}
	p.MakeClockwise()
	p.removeFlatCorners()
	return p
}

// NewPolygonFromArray takes an array of corners and
// reshapes them so that they form a polygon.
func NewPolygonFromArray(corners []*Corner) *Polygon {
	p := &Polygon{corners, len(corners)}
	for i := 0; i < len(corners); i++ {
		p.Corners[(i+1)%p.N].N1 = p.Corners[i]
		p.Corners[i].N2 = p.Corners[(i+1)%p.N]
	}
	p.MakeClockwise()
	p.removeFlatCorners()
	return p
}

// removeFlatCorners removes every corner of the polygon
// that doesn't add to it's shape; that means every corner
// that has an 180 degree angle.
func (P *Polygon) removeFlatCorners() {
	removed := 0
	for i, c := range P.Corners {
		ls := NewLineSegment(*c.N1.Point, *c.N2.Point)
		if ls.L.RelationOf(c.Point) == 0 {
			P.remove(i - removed)
			removed++
		}
	}
}

// remove removes a single corner given by the index
// from a polygon.
func (P *Polygon) remove(indx int) {
	c := P.Corners[indx]
	c.N1.N2 = c.N2
	c.N2.N1 = c.N1
	newCorners := make([]*Corner, P.N-1)
	j := 0
	for i := 0; i < P.N; i++ {
		if i == indx {
			i++
		}
		if j < len(newCorners) && i < P.N {
			newCorners[j] = P.Corners[i]
			j++
		}
	}
	P.Corners = newCorners
	P.N--
}

// AddCorner adds a single corner to the polygon;
// if you are adding more corners at once, please refer
// to (*Polygon).AddCorners.
func (P *Polygon) AddCorner(p *Point) *Polygon {
	if P.N > 0 {
		newCorner := &Corner{p, P, P.Corners[P.N-1], P.Corners[0]}
		P.Corners = append(P.Corners, newCorner)
	} else {
		P.Corners = append(P.Corners, &Corner{p, P, nil, nil})
	}
	P.N++
	P.Corners[P.N-2].N2 = P.Corners[P.N-1]
	P.Corners[0].N1 = P.Corners[P.N-1]
	return P
}

// AddCorners adds a slice of corners to
// the polygon efficiently
func (P *Polygon) AddCorners(Points []*Point) *Polygon {
	newN := P.N + len(Points)
	newCorners := make([]*Corner, newN)
	copy(newCorners[:P.N], P.Corners)
	for i := 0; i < len(Points); i++ {
		newCorners[P.N+i] = &Corner{Points[i], P, nil, nil}
	}
	for i := 0; i < len(Points); i++ {
		newCorners[(P.N+i+1)%newN].N1 = newCorners[P.N+i]
		newCorners[P.N+i].N2 = newCorners[(P.N+i+1)%newN]
	}
	P.Corners = newCorners
	P.N = newN
	P.MakeClockwise()
	return P
}

// SToPlygn converts a string of ints into a polygon as
// implemented by Polygon.
func SToPlygn(s string) *Polygon {
	splitS := strings.Split(s, " ")
	cnt := tools.Stoi(splitS[0])
	corners := make([]*Corner, cnt)
	for i := 0; i < cnt; i++ {
		x := tools.Stof64(splitS[2*i+1])
		y := tools.Stof64(splitS[2*i+2])
		corners[i] = &Corner{NewPoint(x, y), nil, nil, nil}
	}
	return NewPolygonFromArray(corners)
}

// MakeClockwise orders the corners clockwise
func (P *Polygon) MakeClockwise() {
	if !P.isClockwise() {
		P.Reverse()
	}
}

// See https://stackoverflow.com/questions/1165647/how-to-determine-if-a-list-of-polygon-points-are-in-clockwise-order
func (P *Polygon) isClockwise() bool {
	var s float64
	for _, c := range P.Corners {
		if c.N2.Point.Y != -c.Point.Y {
			s += (c.N2.Point.X - c.Point.X) /
				(c.N2.Point.Y + c.Point.Y)
		}
	}
	return s > 0
}

// Reverse changes the order of a polygon
// (clockwise <-> counterclockwise)
func (P *Polygon) Reverse() {
	newCorners := make([]*Corner, P.N)
	for i := 0; i < P.N; i++ {
		newCorners[i] = P.Corners[P.N-i-1]
	}
	for i := 0; i < P.N; i++ {
		newCorners[(i+1)%P.N].N1 = newCorners[i]
		newCorners[i].N2 = newCorners[(i+1)%P.N]
	}
	P.Corners = newCorners
}

// GetBox returns the minimal rectangle that is alligned
// to the axis that includes every point of the polygon
func (P *Polygon) GetBox() (*Point, *Point) {
	if P.N == 0 {
		return &Point{0, 0}, &Point{0, 0}
	} else if P.N == 1 {
		return P.Corners[0].Point, P.Corners[0].Point
	} else {
		min := P.Corners[0]
		max := P.Corners[0]
		for i := 0; i < P.N; i++ {
			if P.Corners[i].Y < min.Y {
				min.Y = P.Corners[i].Y
			}
			if P.Corners[i].Y > max.Y {
				max.Y = P.Corners[i].Y
			}
			if P.Corners[i].X < min.X {
				min.X = P.Corners[i].X
			}
			if P.Corners[i].X > max.X {
				max.X = P.Corners[i].X
			}
		}
		return min.Point, max.Point
	}
}

// HullContains returns wether a given point p is inside the
// convex hull of the polygon
func (P *Polygon) HullContains(p *Point) bool {
	if P.N == 0 {
		// If there are no corners, p can't be inside
		return false
	} else if P.N == 1 {
		// if there is only one corner,
		// p must equal that one
		return *p == *P.Corners[0].Point
	}
	// if it's not inside the rectangle surrounding
	// the polygon, it's not inside the polygon
	min, max := P.GetBox()
	if p.X <= min.X || p.Y <= min.Y ||
		p.X >= max.X || p.Y >= max.Y {
		return false
	}
	// if we draw a line to the left, we can count
	// the intersections; a point inside the polygon
	// intersects odd times
	var intersections int
	ln := NewLineSegment(*p, *NewPoint(max.X+1, p.Y))
	q := P.Corners[0]
	// see below
	for q.IsConcave() {
		q = q.N2
	}
	for i := 0; i < P.N; i++ {
		nextQ := q.N2
		// since we're only concerned with the hull,
		// we don't want to count the concave corners
		for nextQ.IsConcave() {
			i++
			nextQ = nextQ.N2
		}
		edge := NewLineSegment(*q.Point, *nextQ.Point)
		if intersection(ln, edge) < 0 {
			intersections++
		}
		q = nextQ
	}
	return intersections%2 == 1
}

// DoesIntersect returns true if the given line segment
// intersects the polygon
func (P *Polygon) DoesIntersect(ls *LineSegment) bool {
	// Find out wether each corner is above, below or on the line
	relations := make([]int, P.N)
	for i := 0; i < P.N; i++ {
		relations[i] = ls.L.RelationOf(P.Corners[i].Point)
	}
	// Check for intersections in between edges
	for i := 0; i < P.N; i++ {
		if relations[i] == 0 {
			if ls.inCounterClockwiseAngle(P.Corners[i].Point,
				P.Corners[i].N2.Point, P.Corners[i].N1.Point) {
				return true
			}
		}
	}
	// Check for intersections \w edges
	for i := 0; i < P.N; i++ {
		j := (i + 1) % P.N
		if relations[i]*relations[j] < 0 {
			newLS := NewLineSegment(*P.Corners[i].Point, *P.Corners[j].Point)
			if newLS.L.RelationOf(ls.A)*newLS.L.RelationOf(ls.B) < 0 {
				return true
			}
		}
	}
	return false
}

// ErrPointEnclosed indicates that a given point does
// not have a shadow because a concave polygon encloses it
var ErrPointEnclosed = errors.New(
	"The given Polygon is concave and the given point is within it's hull")

// GetShadow returns the corners of the polygon which
// would determine the shadow if there was a light
// source at a given point; the points are ordered counter clockwise
func (P *Polygon) GetShadow(p *Point) (*Corner, *Corner, error) {
	if P.HullContains(p) {
		return nil, nil, ErrPointEnclosed
	}
	// The eventual output
	var p1, p2 *Corner
	// Indicates whether the first point was found
	var found bool
	// Getting the distance but not unsquaring it;
	// this would waste resources;
	// x > y <=> x*x > y*y
	maxSqDist := GetSqDist(p, P.Corners[0].Point)
	var d float64
	// We need to find the furthest distance to
	// check for intersections; if there are none
	// in the radius of this point, there are none
	// at all
	for _, c := range P.Corners {
		d = GetSqDist(p, c.Point)
		if d > maxSqDist {
			maxSqDist = d
		}
	}
	for _, c := range P.Corners {
		if c.Point != p {
			ls := NewLineSegment(*p, *c.Point).ScaleTo(maxSqDist)
			if !P.DoesIntersect(ls) {
				if found {
					p2 = c
					break
				}
				p1 = c
				found = true
			}
		}
	}
	// Getting angles to sort the points
	angle1, err := GetAngle(p, p1.Point)
	if err != nil {
		panic(err)
	}
	angle2, err := GetAngle(p, p2.Point)
	if err != nil {
		panic(err)
	}
	// Getting a reference point to check
	// which one of the points comes first in
	// counter clockwise order
	var refP *Point
	for _, c := range P.Corners {
		if c.Point != p1.Point && c.Point != p2.Point {
			refP = c.Point
			break
		}
	}
	// Getting reference angle
	refAngle, err := GetAngle(p, refP)
	if err != nil {
		refAngle = 0
	}
	// If they are in the wrong order, swap them
	if ToPositiveAngle(refAngle-angle1) <
		ToPositiveAngle(refAngle-angle2) {
		p1, p2 = p2, p1
	}
	// Finally, return them
	return p1, p2, nil
}

// GetUnblockedSpikes returns every convex corner of
// the polygon that can be connected to p with a straight
// line without intersecting the edges of the polygon
func (P *Polygon) GetUnblockedSpikes(p *Point) []*Point {
	res := make([]*Point, 0)
	for _, c := range P.Corners {
		if !c.IsConcave() {
			if !P.DoesIntersect(NewLineSegment(*p, *c.Point)) {
				res = append(res, c.Point)
			}
		}
	}
	return res
}

// contains is the same as hullContains without
// the checks for concave/non-concave
func (P *Polygon) contains(p *Point) bool {
	if P.N == 0 {
		// If there are no corners, p can't be inside
		return false
	} else if P.N == 1 {
		// if there is only one corner,
		// p must equal that one
		return p == P.Corners[0].Point
	}
	// if it's not inside the rectangle surrounding
	// the polygon, it's not inside the polygon
	min, max := P.GetBox()
	if p.X <= min.X || p.Y <= min.Y ||
		p.X >= max.X || p.Y >= max.Y {
		return false
	}
	// if we draw a line to the left, we can count
	// the intersections; a point inside the polygon
	// intersects odd times
	var intersections int
	ln := NewLineSegment(*p, *NewPoint(max.X+1, p.Y))
	q := P.Corners[0]
	for i := 0; i < P.N; i++ {
		nextQ := q.N2
		edge := NewLineSegment(*q.Point, *nextQ.Point)
		if intersection(ln, edge) < 0 {
			intersections++
		}
		q = nextQ
	}
	return intersections%2 == 1
}
