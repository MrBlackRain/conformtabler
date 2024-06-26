local function deepcopy(orig)
	local orig_type = type(orig)
	local copy
	if orig_type == "table" then
		copy = {}
		for orig_key, orig_value in next, orig, nil do
			copy[deepcopy(orig_key)] = deepcopy(orig_value)
		end
		setmetatable(copy, deepcopy(getmetatable(orig)))
	else
		copy = orig
	end
	return copy
end

local uv = {
	os_uname = function()
		return {
			version = "",
			sysname = "",
		}
	end,
}

vim = { deepcopy = deepcopy, uv = uv }
