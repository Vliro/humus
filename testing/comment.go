package gen
//Save overrides default save behavior.
func (c *Comment) Save() {
	mv := c.MapValues()
	mv.Set(CommentCommentsOnField, true, c.CommentsOn)
}
