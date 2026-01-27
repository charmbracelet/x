#!/usr/bin/env lua
-- Extract LSP configurations from nvim-lspconfig and output JSON.

local lsp_dir = arg[1]
if not lsp_dir then
    io.stderr:write("Usage: lua extract-lsp-configs.lua <lsp-dir>\n")
    os.exit(1)
end

-- Minimal vim mock - just enough to load config files without errors.
local noop = function() end
local empty = function() return {} end
local ret_nil = function() return nil end
local ret_str = function() return "" end

_G.vim = setmetatable({
    fn = setmetatable({}, { __index = function() return ret_str end }),
    api = setmetatable({}, { __index = function() return empty end }),
    env = setmetatable({}, { __index = function() return "" end }),
    fs = { root = ret_nil, dirname = ret_str, joinpath = function(...) return table.concat({...}, "/") end, normalize = function(p) return p or "" end, find = empty },
    uv = { os_homedir = function() return os.getenv("HOME") or "/home/user" end, os_tmpdir = function() return "/tmp" end, fs_stat = ret_nil, cwd = function() return "/" end },
    lsp = { get_clients = empty, rpc = { connect = function(h, p) return { host = h, port = p } end }, protocol = { make_client_capabilities = empty }, config = setmetatable({}, { __index = empty, __call = empty }) },
    g = setmetatable({}, { __index = ret_nil, __newindex = noop }),
    o = setmetatable({}, { __index = ret_str }),
    bo = setmetatable({}, { __index = ret_str }),
    log = { levels = { DEBUG = 1, INFO = 2, WARN = 3, ERROR = 4 } },
    diagnostic = { severity = { ERROR = 1, WARN = 2, INFO = 3, HINT = 4 } },
    tbl_deep_extend = function(_, ...) local r = {} for _, t in ipairs({...}) do if type(t) == "table" then for k, v in pairs(t) do r[k] = v end end end return r end,
    tbl_extend = function(_, ...) local r = {} for _, t in ipairs({...}) do if type(t) == "table" then for k, v in pairs(t) do r[k] = v end end end return r end,
    list_extend = function(dst, src) for _, v in ipairs(src) do dst[#dst+1] = v end return dst end,
    deepcopy = function(t) if type(t) ~= "table" then return t end local c = {} for k, v in pairs(t) do c[k] = _G.vim.deepcopy(v) end return c end,
    split = function(s, sep) local r = {} for m in (s..sep):gmatch("(.-)"..sep) do r[#r+1] = m end return r end,
    trim = function(s) return type(s) == "string" and s:match("^%s*(.-)%s*$") or "" end,
    startswith = function(s, p) return type(s) == "string" and s:sub(1, #p) == p end,
    notify = noop, schedule = noop, system = noop, cmd = noop, inspect = tostring,
    uri_from_bufnr = ret_str, uri_from_fname = function(f) return "file://" .. f end,
    empty_dict = empty, deprecate = noop, is_callable = function() return false end,
    version = function() return { major = 0, minor = 11, patch = 0 } end,
    filetype = { match = ret_nil },
}, { __index = function() return noop end })

-- Flatten nested tables (like lspconfig's tbl_flatten)
local function tbl_flatten(t)
    local result = {}
    local function flatten(tbl)
        for _, v in ipairs(tbl) do
            if type(v) == "table" then flatten(v) else result[#result+1] = v end
        end
    end
    flatten(t)
    return result
end

-- Create a callable table that stores patterns and can be identified later
local function make_root_pattern(...)
    local patterns = tbl_flatten({...})
    -- Return a callable table that stores the patterns
    local obj = { _root_patterns = patterns }
    setmetatable(obj, { __call = function() return nil end })
    return obj
end

package.preload['lspconfig.util'] = function()
    return {
        root_pattern = make_root_pattern,
        find_git_ancestor = ret_nil, find_node_modules_ancestor = ret_nil, find_package_json_ancestor = ret_nil,
        insert_package_json = noop,
        path = { exists = function() return false end, join = function(...) return table.concat({...}, "/") end, is_dir = function() return false end, is_file = function() return false end, dirname = ret_str, is_absolute = function() return true end },
        get_active_clients_list_by_ft = empty, get_active_client_by_name = ret_nil,
        tbl_flatten = tbl_flatten,
    }
end
package.preload['lspconfig/util'] = package.preload['lspconfig.util']
package.preload['lspconfig.async'] = function() return { run = noop } end
package.preload['lspconfig'] = function() return { util = package.preload['lspconfig.util']() } end

-- JSON encoder
local function json(val)
    local t = type(val)
    if t == "nil" then return "null"
    elseif t == "boolean" then return val and "true" or "false"
    elseif t == "number" then return tostring(val)
    elseif t == "string" then return '"' .. val:gsub('\\', '\\\\'):gsub('"', '\\"'):gsub('\n', '\\n'):gsub('\r', '\\r'):gsub('\t', '\\t') .. '"'
    elseif t == "table" then
        local is_arr, max = true, 0
        for k in pairs(val) do
            if type(k) ~= "number" or k < 1 or k ~= math.floor(k) then is_arr = false; break end
            if k > max then max = k end
        end
        if is_arr and max > 0 then for i = 1, max do if val[i] == nil then is_arr = false; break end end end
        if max == 0 then is_arr = false end
        if is_arr then
            local parts = {} for i = 1, max do parts[i] = json(val[i]) end
            return "[" .. table.concat(parts, ", ") .. "]"
        else
            local parts = {} for k, v in pairs(val) do if type(k) == "string" then parts[#parts+1] = json(k) .. ": " .. json(v) end end
            table.sort(parts)
            return "{" .. table.concat(parts, ", ") .. "}"
        end
    end
end

-- Check if value is JSON-serializable (no functions)
local function is_serializable(val, depth)
    if (depth or 0) > 10 then return false end
    local t = type(val)
    if t == "function" then return false end
    if t ~= "table" then return true end
    for k, v in pairs(val) do
        if type(k) ~= "string" and type(k) ~= "number" then return false end
        if not is_serializable(v, (depth or 0) + 1) then return false end
    end
    return true
end

-- Extract root patterns from source text (fallback for function-wrapped root_dir)
local function extract_patterns_from_source(source)
    local patterns = {}
    local seen = {}
    
    -- Match root_pattern('a', 'b', ...) or root_pattern { 'a', 'b' }
    for args in source:gmatch("root_pattern%s*[%({]([^%)%}]+)[%)%}]") do
        for pattern in args:gmatch("['\"]([^'\"]+)['\"]") do
            if not seen[pattern] then
                patterns[#patterns+1] = pattern
                seen[pattern] = true
            end
        end
    end
    
    -- Match vim.fs.root(fname, 'pattern') or vim.fs.root(0, 'pattern')
    for pattern in source:gmatch("vim%.fs%.root%s*%([^,]+,%s*['\"]([^'\"]+)['\"]%s*%)") do
        if not seen[pattern] then
            patterns[#patterns+1] = pattern
            seen[pattern] = true
        end
    end
    
    return patterns
end

-- Extract fields we care about from a config
local ignore_keys = { editorInfo = true, editorPluginInfo = true }
local home_dir = os.getenv("HOME") or "/home/user"

local function filter_table(t)
    if type(t) ~= "table" then
        -- Skip stringified table references (e.g., "table: 0x...")
        if type(t) == "string" and t:match("^table: 0x") then
            return nil
        end
        -- Replace home directory with $HOME in strings
        if type(t) == "string" and t:find(home_dir, 1, true) then
            return t:gsub(home_dir, "$HOME")
        end
        return t
    end
    local r = {}
    for k, v in pairs(t) do
        if not ignore_keys[k] then
            local filtered = filter_table(v)
            if filtered ~= nil then r[k] = filtered end
        end
    end
    return next(r) and r or nil
end

local function extract(cfg, source)
    if not cfg.cmd then return nil end
    local r = {}
    
    if type(cfg.cmd) == "table" and #cfg.cmd > 0 then
        r.command = cfg.cmd[1]
        if #cfg.cmd > 1 then r.args = {} for i = 2, #cfg.cmd do r.args[#r.args+1] = cfg.cmd[i] end end
    elseif type(cfg.cmd) == "string" then
        r.command = cfg.cmd
    end
    if not r.command then return nil end
    
    if type(cfg.filetypes) == "table" then
        local ft = {} for _, v in ipairs(cfg.filetypes) do if type(v) == "string" then ft[#ft+1] = v end end
        if #ft > 0 then r.filetypes = ft end
    end
    
    -- Check for root_markers (explicit) or extract from root_dir (root_pattern result)
    local markers = {}
    if type(cfg.root_markers) == "table" then
        for _, v in ipairs(cfg.root_markers) do if type(v) == "string" then markers[#markers+1] = v end end
    end
    -- Extract patterns from root_dir if it's a root_pattern result
    if type(cfg.root_dir) == "table" and cfg.root_dir._root_patterns then
        for _, v in ipairs(cfg.root_dir._root_patterns) do
            if type(v) == "string" then markers[#markers+1] = v end
        end
    end
    -- Fallback: parse source for patterns if root_dir is a function
    if #markers == 0 and type(cfg.root_dir) == "function" and source then
        markers = extract_patterns_from_source(source)
    end
    if #markers > 0 then r.root_markers = markers end
    
    if type(cfg.settings) == "table" and is_serializable(cfg.settings) then
        local filtered = filter_table(cfg.settings)
        if filtered then r.settings = filtered end
    end
    if type(cfg.init_options) == "table" and is_serializable(cfg.init_options) then
        local filtered = filter_table(cfg.init_options)
        if filtered then r.init_options = filtered end
    end
    
    return r
end

-- Main
local configs, names = {}, {}
local handle = io.popen('ls "' .. lsp_dir .. '"/*.lua 2>/dev/null')
if handle then
    for filepath in handle:lines() do
        local name = filepath:match("([^/]+)%.lua$")
        if name then
            -- Read source for pattern extraction fallback
            local f = io.open(filepath, "r")
            local source = f and f:read("*a") or ""
            if f then f:close() end
            
            local chunk = loadfile(filepath)
            if chunk then
                local ok, result = pcall(chunk)
                if ok and type(result) == "table" then
                    local cfg = extract(result.default_config or result, source)
                    if cfg then configs[name] = cfg; names[#names+1] = name end
                end
            end
        end
    end
    handle:close()
end

table.sort(names)
io.write("{\n")
for i, name in ipairs(names) do
    io.write('  "' .. name .. '": ' .. json(configs[name]) .. (i < #names and "," or "") .. "\n")
end
io.write("}\n")
io.stderr:write("Generated " .. #names .. " server configs\n")
