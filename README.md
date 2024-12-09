代码仓库：https://gitcode.com/xu1feng/hm-dianpnig/overview

# 整体功能架构图

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411052029966.png)

# 短信登录

## 导入黑马点评项目

首先，导入数据库SQL文件`hmdp.sql`。

其中的表有：

- tb_user：用户表
- tb_user_info：用户详情表
- tb_shop：商户信息表
- tb_shop_type：商户类型表
- tb_blog：用户日记表（达人探店日记）
- tb_follow：用户关注表
- tb_voucher：优惠券表
- tb_voucher_order：优惠券的订单表

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411052056176.png)

将资料中的源码`hm-dianping`导入到idea中，启动项目后，在浏览器访问：[http://localhost:8001/shop-type/list](http://localhost:8001/shop-type/list)，如果可以看到数据则证明运行没有问题。

**注意：**记得修改application.yml中的mysql、redis地址信息

将资料中的Nginx复制到英文目录下，并双击nginx.exe，然后在chrome浏览器中打开手机模式
![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411052130576.png)

然后访问http://localhost:8080/，即可看到页面
![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411052136244.png)

## 基于Session实现登录

### 发送短信验证码

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411052139443.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411052146061.png)

在UserController.java中编写如下代码

```java
/**  
 * 发送手机验证码  
 */  
@PostMapping("code")  
public Result sendCode(@RequestParam("phone") String phone, HttpSession session) {  
    return userService.sendCode(phone, session);  
}
```

在IUserService.java中编写如下代码

```java
Result sendCode(String phone, HttpSession session);
```

在UserServiceImpl.java中编写如下代码

```java
@Slf4j  
@Service  
public class UserServiceImpl extends ServiceImpl<UserMapper, User> implements IUserService {  
  
    @Override  
    public Result sendCode(String phone, HttpSession session) {  
        // 1.校验手机号  
        if (RegexUtils.isPhoneInvalid(phone)) {  
            // 2.如果不符合，返回错误信息  
            return Result.fail("手机号格式错误");  
        }  
  
        // 3.符合，生成验证码  
        String code = RandomUtil.randomNumbers(6);  
  
        // 4.保存验证码到session  
        session.setAttribute("code", code);  
  
        // 5.发送验证码  
        log.debug("发送短信验证码成功，验证码：{}", code);  
  
        // 返回ok  
        return Result.ok();  
    }  
}
```

### 短信验证码登录、注册

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411052140683.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411060900655.png)

在UserController.java中编写

```java
/**  
 * 登录功能  
 * @param loginForm 登录参数，包含手机号、验证码；或者手机号、密码  
 */  
@PostMapping("/login")  
public Result login(@RequestBody LoginFormDTO loginForm, HttpSession session){  
    // 实现登录功能  
    return userService.login(loginForm, session);  
}
```

在IUserService.java中编写

```java
public interface IUserService extends IService<User> {  
  
    Result sendCode(String phone, HttpSession session);  
  
    Result login(LoginFormDTO loginForm, HttpSession session);  
}
```

在UserServiceImpl.java中编写

```java
@Override  
public Result login(LoginFormDTO loginForm, HttpSession session) {  
    // 1.校验手机号  
    String phone = loginForm.getPhone();  
    if (RegexUtils.isPhoneInvalid(phone)) {  
        // 2.如果不符合，返回错误信息  
        return Result.fail("手机号格式错误");  
    }  
  
    // 2.校验验证码  
    Object cacheCode = session.getAttribute("code");  
    String code = loginForm.getCode();  
    if (cacheCode == null || !cacheCode.toString().equals(code)) {  
        // 3.不一致，报错  
        return Result.fail("验证码错误");  
    }  
  
    // 4.一致，根据手机号查询用户 select * from tb_user where phone = ?    User user = query().eq("phone", phone).one();  
  
    // 5.判断用户是否存在  
    if (user == null) {  
        // 6.不存在，创建新用户并保存  
        user = createUserWithPhone(phone);  
    }  
  
    // 7.保存用户信息到session  
    session.setAttribute("user", user);  
    return Result.ok();  
}  
  
private User createUserWithPhone(String phone) {  
    // 1.创建用户  
    User user = new User();  
    user.setPhone(phone);  
    user.setNickName(USER_NICK_NAME_PREFIX + RandomUtil.randomString(10));  
  
    // 2.保存用户  
    save(user);  
    return user;  
}
```

### 校验登录状态

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/202411052141478.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241111202400.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241111202507.png)

在utils包下创建登录校验拦截器LoginInterceptor.java

```java
public class LoginInterceptor implements HandlerInterceptor {  
  
    @Override  
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {  
        // 1.获取session  
        HttpSession session = request.getSession();  
  
        // 2.获取session中的用户  
        Object user = session.getAttribute("user");  
  
        // 3.判断用户是否存在  
        if (user == null) {  
            // 4.不存在，拦截  返回401状态码  
            response.setStatus(401);  
            return false;  
        }  
  
        // 5.存在，保存用户信息到ThreadLocal  
        UserHolder.saveUser((UserDTO) user);  
  
        // 6.放行  
        return false;  
    }  
  
    @Override  
    public void afterCompletion(HttpServletRequest request, HttpServletResponse response, Object handler, Exception ex) throws Exception {  
        // 移除用户  
        UserHolder.removeUser();  
    }  
}
```

在UserController中修改代码

```java
@GetMapping("/me")  
public Result me(){  
    // 获取当前登录的用户并返回  
    UserDTO user = UserHolder.getUser();  
    return Result.ok(user);  
}
```

在config包下配置添加拦截器代码MvcConfig.java

```java
@Configuration  
public class MvcConfig implements WebMvcConfigurer {  
  
    @Override  
    public void addInterceptors(InterceptorRegistry registry) {  
        registry.addInterceptor(new LoginInterceptor())  
                .excludePathPatterns(  
                        "/shop/**",  
                        "/voucher/**",  
                        "/shop-type/**",  
                        "/upload/**",  
                        "/blog/hot",  
                        "/user/code",  
                        "/user/login"  
                );  
    }  
}
```

## 集群的session共享问题

**session共享问题**：多台Tomcat并不共享session存储空间，当请求切换到不同tomcat服务时导致数据丢失的问题。

session的替代方案应该满足：

- 数据共享
- 内存存储
- key、value结构

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241111212836.png)

## 基于Redis实现共享session登录

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241111213531.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241111214043.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241111214349.png)

UserServiceImpl.java

```java
@Slf4j  
@Service  
public class UserServiceImpl extends ServiceImpl<UserMapper, User> implements IUserService {  
  
    @Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    @Override  
    public Result sendCode(String phone, HttpSession session) {  
        // 1.校验手机号  
        if (RegexUtils.isPhoneInvalid(phone)) {  
            // 2.如果不符合，返回错误信息  
            return Result.fail("手机号格式错误");  
        }  
  
        // 3.符合，生成验证码  
        String code = RandomUtil.randomNumbers(6);  
  
        // 4.保存验证码到 redis  set key value ex 120        stringRedisTemplate.opsForValue().set(RedisConstants.LOGIN_CODE_KEY + phone, code, RedisConstants.LOGIN_CODE_TTL, TimeUnit.MINUTES);  
  
        // 5.发送验证码  
        log.debug("发送短信验证码成功，验证码：{}", code);  
  
        // 返回ok  
        return Result.ok();  
    }  
  
    @Override  
    public Result login(LoginFormDTO loginForm, HttpSession session) {  
        // 1.校验手机号  
        String phone = loginForm.getPhone();  
        if (RegexUtils.isPhoneInvalid(phone)) {  
            // 2.如果不符合，返回错误信息  
            return Result.fail("手机号格式错误");  
        }  
  
        // 2.从redis获取验证码并校验  
        String cacheCode = stringRedisTemplate.opsForValue().get(RedisConstants.LOGIN_CODE_KEY + phone);  
        String code = loginForm.getCode();  
        if (cacheCode == null || !cacheCode.toString().equals(code)) {  
            // 3.不一致，报错  
            return Result.fail("验证码错误");  
        }  
  
        // 4.一致，根据手机号查询用户 select * from tb_user where phone = ?        User user = query().eq("phone", phone).one();  
  
        // 5.判断用户是否存在  
        if (user == null) {  
            // 6.不存在，创建新用户并保存  
            user = createUserWithPhone(phone);  
        }  
  
        // 7.保存用户信息到session  
        // 7.1 随机生成token，作为登录令牌  
        String token = UUID.randomUUID().toString(true);  
  
        // 7.2 将User对象转为HashMap存储  
        UserDTO userDTO = BeanUtil.copyProperties(user, UserDTO.class);  
        Map<String, Object> userMap = BeanUtil.beanToMap(userDTO, new HashMap<>(), CopyOptions.create()  
                .setIgnoreNullValue(true)  
                .setFieldValueEditor((filedName, fieldValue) -> fieldValue.toString()));  
  
        // 7.3 存储  
        stringRedisTemplate.opsForHash().putAll(RedisConstants.LOGIN_USER_KEY + token, userMap);  
  
        // 7.4 设置token有效期  
        stringRedisTemplate.expire(RedisConstants.LOGIN_USER_KEY + token, RedisConstants.LOGIN_USER_TTL, TimeUnit.MINUTES);  
  
        // 8. 返回token  
        return Result.ok(token);  
    }  
  
    private User createUserWithPhone(String phone) {  
        // 1.创建用户  
        User user = new User();  
        user.setPhone(phone);  
        user.setNickName(SystemConstants.USER_NICK_NAME_PREFIX + RandomUtil.randomString(10));  
  
        // 2.保存用户  
        save(user);  
        return user;  
    }  
}
```

LoginInterceptor.java

```java
public class LoginInterceptor implements HandlerInterceptor {  
  
    private StringRedisTemplate stringRedisTemplate;  
  
    public LoginInterceptor(StringRedisTemplate stringRedisTemplate) {  
        this.stringRedisTemplate = stringRedisTemplate;  
    }  
  
    @Override  
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {  
        // 1.获取请求头中的token  
        String token = request.getHeader("authorization");  
        if (StrUtil.isBlank(token)) {  
            // 4.不存在，拦截  返回401状态码  
            response.setStatus(401);  
            return false;  
        }  
  
        // 2.基于token获取redis中的用户  
        Map<Object, Object> userMap = stringRedisTemplate.opsForHash().entries(RedisConstants.LOGIN_USER_KEY + token);  
  
        // 3.判断用户是否存在  
        if (userMap.isEmpty()) {  
            // 4.不存在，拦截  返回401状态码  
            response.setStatus(401);  
            return false;  
        }  
  
        // 5.将查询到的Hash数据再转为UserDTO对象  
        UserDTO userDTO = BeanUtil.fillBeanWithMap(userMap, new UserDTO(), false);  
  
        // 6.存在，保存用户信息到ThreadLocal  
        UserHolder.saveUser(userDTO);  
  
        // 7.刷新token有效期  
        stringRedisTemplate.expire(RedisConstants.LOGIN_USER_KEY + token, RedisConstants.LOGIN_USER_TTL, TimeUnit.MINUTES);  
  
        // 8.放行  
        return false;  
    }  
  
    @Override  
    public void afterCompletion(HttpServletRequest request, HttpServletResponse response, Object handler, Exception ex) throws Exception {  
        // 移除用户  
        UserHolder.removeUser();  
    }  
}
```

MvcConfig.java

```java
@Configuration  
public class MvcConfig implements WebMvcConfigurer {  
  
    @Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    @Override  
    public void addInterceptors(InterceptorRegistry registry) {  
        registry.addInterceptor(new LoginInterceptor(stringRedisTemplate))  
                .excludePathPatterns(  
                        "/shop/**",  
                        "/voucher/**",  
                        "/shop-type/**",  
                        "/upload/**",  
                        "/blog/hot",  
                        "/user/code",  
                        "/user/login"  
                );  
    }  
}
```

### 登录拦截器的优化

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113185900.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113185956.png)

在utils包下写RefreshTokenInterceptor.java

```java
public class RefreshTokenInterceptor implements HandlerInterceptor {  
  
    private StringRedisTemplate stringRedisTemplate;  
  
    public RefreshTokenInterceptor(StringRedisTemplate stringRedisTemplate) {  
        this.stringRedisTemplate = stringRedisTemplate;  
    }  
  
    @Override  
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {  
        // 1.获取请求头中的token  
        String token = request.getHeader("authorization");  
        if (StrUtil.isBlank(token)) {  
            return true;  
        }  
  
        // 2.基于token获取redis中的用户  
        Map<Object, Object> userMap = stringRedisTemplate.opsForHash().entries(RedisConstants.LOGIN_USER_KEY + token);  
  
        // 3.判断用户是否存在  
        if (userMap.isEmpty()) {  
            return true;  
        }  
  
        // 5.将查询到的Hash数据再转为UserDTO对象  
        UserDTO userDTO = BeanUtil.fillBeanWithMap(userMap, new UserDTO(), false);  
  
        // 6.存在，保存用户信息到ThreadLocal  
        UserHolder.saveUser(userDTO);  
  
        // 7.刷新token有效期  
        stringRedisTemplate.expire(RedisConstants.LOGIN_USER_KEY + token, RedisConstants.LOGIN_USER_TTL, TimeUnit.MINUTES);  
  
        // 8.放行  
        return false;  
    }  
  
    @Override  
    public void afterCompletion(HttpServletRequest request, HttpServletResponse response, Object handler, Exception ex) throws Exception {  
        // 移除用户  
        UserHolder.removeUser();  
    }  
}
```

LoginInterceptor.java

```java
public class LoginInterceptor implements HandlerInterceptor {  
  
    @Override  
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {  
        // 1.判断是否需要拦截（ThreadLocal中是否有用户）  
        if (UserHolder.getUser() == null) {  
            // 没有，需要拦截，设置状态码  
            response.setStatus(401);  
            // 拦截  
            return false;  
        }  
        // 有用户，放行  
        return false;  
    }  
  
}
```

MvcConfig.java

```java
@Configuration  
public class MvcConfig implements WebMvcConfigurer {  
  
    @Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    @Override  
    public void addInterceptors(InterceptorRegistry registry) {  
        // 登录拦截器  
        registry.addInterceptor(new LoginInterceptor())  
                .excludePathPatterns(  
                        "/shop/**",  
                        "/voucher/**",  
                        "/shop-type/**",  
                        "/upload/**",  
                        "/blog/hot",  
                        "/user/code",  
                        "/user/login"  
                ).order(1);  
        // token刷新的拦截器  
        registry.addInterceptor(new RefreshTokenInterceptor(stringRedisTemplate)).addPathPatterns("/**").order(0);  
    }  
}
```

# 商品查询缓存

## 什么是缓存

缓存就是数据交换的缓冲区（称作Cache），是存贮数据的临时地方，一般**读写性能较高**。

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113192652.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113193155.png)

## 添加Redis缓存

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113193506.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113193858.png)

ShopController.java

```java
/**  
 * 根据id查询商铺信息  
 * @param id 商铺id  
 * @return 商铺详情数据  
 */  
@GetMapping("/{id}")  
public Result queryShopById(@PathVariable("id") Long id) {  
    return shopService.queryById(id);  
}
```

IShopService.java

```java
public interface IShopService extends IService<Shop> {  
  
    Result queryById(Long id);  
}
```

ShopServiceImpl.java

```java
@Service  
public class ShopServiceImpl extends ServiceImpl<ShopMapper, Shop> implements IShopService {  
  
    @Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    @Override  
    public Result queryById(Long id) {  
        // 1.从redis查询商铺缓存  
        String shopJson = stringRedisTemplate.opsForValue().get(RedisConstants.CACHE_SHOP_KEY + id);  
  
        // 2.判断是否存在  
        if (StrUtil.isNotBlank(shopJson)) {  
            // 3.存在，直接返回  
            Shop shop = JSONUtil.toBean(shopJson, Shop.class);  
            return Result.ok(shop);  
        }  
  
        // 4.不存在，根据id查询数据库  
        Shop shop = getById(id);  
  
        // 5.不存在，返回错误  
        if (shop == null) {  
            return Result.fail("店铺不存在！");  
        }  
  
        // 6.存在，写入redis  
        stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, JSONUtil.toJsonStr(shop));  
  
        // 7.返回  
        return Result.ok(shop);  
    }  
}
```

### 作业： 给店铺类型查询添加缓存

ShopTypeController.java

```java
@GetMapping("list")  
public Result queryTypeList() {  
    List<ShopType> typeList = typeService.queryTypeList();  
    return Result.ok(typeList);  
}
```

IShopTypeService.java

```java
public interface IShopTypeService extends IService<ShopType> {  
  
    List<ShopType> queryTypeList();  
}
```

ShopTypeServiceImpl.java

```java
@Service  
public class ShopTypeServiceImpl extends ServiceImpl<ShopTypeMapper, ShopType> implements IShopTypeService {  
  
    @Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    @Override  
    public List<ShopType> queryTypeList() {  
        String key = RedisConstants.CACHE_TYPE_KEY;  
        // 1.从redis中查询店铺类型  
        String shopTypeJson = stringRedisTemplate.opsForValue().get(key);  
  
        // 2.判断是否存在  
        if (StrUtil.isNotBlank(shopTypeJson)) {  
            // 3.存在，转化为list返回  
            return JSONUtil.toList(shopTypeJson, ShopType.class);  
        }  
  
        // 4.不存在，查询数据库  
        List<ShopType> typeList = this.query().orderByAsc("sort").list();  
        // 5.不存在，返回空列表  
        if (typeList.isEmpty()) {  
            return new ArrayList<>();  
        }  
  
        // 6.存在，写入redis  
        stringRedisTemplate.opsForValue().set(key, JSONUtil.toJsonStr(typeList));  
  
        // 7.返回  
        return typeList;  
    }  
}
```

## 缓存更新策略

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113201245.png)

### 主动更新策略

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113201344.png)

操作缓存和数据库时有三个问题需要考虑：

1. 删除缓存还是更新缓存？
   - 更新缓存：每次更新数据库都更新缓存，无效写操作较多 （❌）
   - 删除缓存：更新数据库时让缓存失效，查询时再更新缓存 （✅）
2. 如何保证缓存与数据库的操作的同时成功或失败？
   - 单体系统，将缓存与数据库操作放在一个事务
   - 分布式系统，利用TCC等分布式事务方案
3. 先操作缓存还是先操作数据库？
   - 先删除缓存，再操作数据库
   - 先操作数据库，再删除缓存

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113203129.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113203159.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113203247.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113203453.png)

缓存更新策略的最佳实践方案：

1. 低一致性需求：使用Redis自带的内存淘汰机制
2. 高一致性需求：主动更新，并以超时剔除作为兜底方案
   - 读操作
     - 缓存命中直接返回
     - 缓存未命中则查询数据库，并写入缓存，设定超时时间
   - 写操作：
     - 先写数据库，然后再删除缓存
     - 要确保数据库与缓存操作的原子性

#### 案例：给查询商铺的缓存添加超时剔除和主动更新的策略

修改ShopController中的业务逻辑，满足下面的需求：

1. 根据id查询店铺时，如果缓存未命中，则查询数据库，将数据库结果写入缓存，并设置超时时间
2. 根据id修改店铺时，先修改数据库，再删除缓存

ShopController.java

```java
@PutMapping  
public Result updateShop(@RequestBody Shop shop) {  
    // 写入数据库  
    return shopService.update(shop);  
}
```

IShopService.java

```java
Result update(Shop shop);
```

ShopServiceImpl.java

```java
@Override  
@Transactional  
public Result update(Shop shop) {  
    Long id = shop.getId();  
    if (id == null) {  
        return Result.fail("店铺id不能为空！");  
    }  
    // 1.更新数据库  
    updateById(shop);  
  
    // 2.删除缓存  
    stringRedisTemplate.delete(RedisConstants.CACHE_SHOP_KEY + id);  
  
    return Result.ok();  
}
```

## 缓存穿透

**缓存穿透**是指客户端请求的数据在缓存中和数据库中都不存在，这样缓存永远不会生效，**这些请求都会打到数据库**。

常见的解决方案有两种：

- 缓存空对象
  - 优点：实现简单，维护方便
  - 缺点：
    - 额外的内存消耗
    - 可能造成短期的不一致
- 布隆过滤
  - 优点：内存占用较少，没有多余key
  - 缺点：
    - 实现复杂
    - 存在误判可能

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113211935.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113212702.png)

ShopServiceImpl.java

```java
@Override  
public Result queryById(Long id) {  
    // 1.从redis查询商铺缓存  
    String shopJson = stringRedisTemplate.opsForValue().get(RedisConstants.CACHE_SHOP_KEY + id);  
  
    // 2.判断是否存在  
    if (StrUtil.isNotBlank(shopJson)) {  
        // 3.存在，直接返回  
        Shop shop = JSONUtil.toBean(shopJson, Shop.class);  
        return Result.ok(shop);  
    }  
  
    // 判断命中的是否是空值  
    if (shopJson != null) {  
        // 返回一个错误信息  
        return Result.fail("店铺信息不存在！");  
    }  
  
    // 4.不存在，根据id查询数据库  
    Shop shop = getById(id);  
  
    // 5.不存在，返回错误  
    if (shop == null) {  
        // 将空值写入redis  
        stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, "", RedisConstants.CACHE_NULL_TTL, TimeUnit.MINUTES);  
        // 返回错误信息  
        return Result.fail("店铺不存在！");  
    }  
  
    // 6.存在，写入redis  
    stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, JSONUtil.toJsonStr(shop), RedisConstants.CACHE_SHOP_TTL, TimeUnit.MINUTES);  
  
    // 7.返回  
    return Result.ok(shop);  
}
```

缓存穿透产生的原因：

- 用户请求的数据在缓存中和数据库中都不存在，不断发起这样的请求，给数据库带来巨大压力
  缓存穿透的解决方案：
- 缓存null值
- 布隆过滤
- 增强id的复杂度，避免被猜测id规律
- 做好数据的基础格式校验
- 加强用户权限校验
- 做好热点参数的限流

## 缓存雪崩

**缓存雪崩**是指在同一时段大量的缓存key同时失效或者Redis服务宕机，导致大量请求到达数据库，带来巨大压力。

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113225053.png)
**解决方案：**

- 给不同的Key的TTL添加随机值
- 利用Redis集群提高服务的可用性
- 给缓存业务添加降级限流策略
- 给业务添加多级缓存

## 缓存击穿

**缓存击穿问题**也叫热点Key问题，就是一个被**高并发访问**并且**缓存重建业务较复杂**的key突然失效了，无数的请求访问会在瞬间给数据库带来巨大的冲击。
常见的解决方案有两种：

- 互斥锁
- 逻辑过期

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113230635.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113231437.png)

![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113231735.png)

#### 案例：基于互斥锁方式解决缓存击穿问题

需求：修改根据id查询商铺的业务，基于互斥锁方式来解决缓存击穿问题
![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241113232551.png)

ShopServiceImpl.java

```java
@Service  
public class ShopServiceImpl extends ServiceImpl<ShopMapper, Shop> implements IShopService {  
  
    @Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    @Override  
    public Result queryById(Long id) {  
        // 缓存穿透  
//        Shop shop = queryWithPassThrough(id);  
  
        // 互斥锁解决缓存击穿  
        Shop shop = queryWithMutex(id);  
        if (shop == null) {  
            return Result.fail("店铺不存在！");  
        }  
        return Result.ok(shop);  
    }  
  
    private Shop queryWithMutex(Long id) {  
        // 1.从redis查询商铺缓存  
        String shopJson = stringRedisTemplate.opsForValue().get(RedisConstants.CACHE_SHOP_KEY + id);  
  
        // 2.判断是否存在  
        if (StrUtil.isNotBlank(shopJson)) {  
            // 3.存在，直接返回  
            return JSONUtil.toBean(shopJson, Shop.class);  
        }  
  
        // 判断命中的是否是空值  
        if (shopJson != null) {  
            // 返回一个错误信息  
            return null;  
        }  
  
        // 4.开始实现缓存重建  
        // 4.1获取互斥锁  
        String lockKey = "lock:shop:" + id;  
        Shop shop = null;  
        try {  
            boolean isLock = tryLock(lockKey);  
  
            // 4.2判断是否获取成功  
            if (!isLock) {  
                // 4.3失败，休眠并重试  
                Thread.sleep(50);  
                return queryWithMutex(id);  
            }  
  
            // 4.4成功，根据id查询数据库  
            shop = getById(id);  
  
            // 模拟重建延时  
            Thread.sleep(200);  
  
            // 5.不存在，返回错误  
            if (shop == null) {  
                // 将空值写入redis  
                stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, "", RedisConstants.CACHE_NULL_TTL, TimeUnit.MINUTES);  
                // 返回错误信息  
                return null;  
            }  
  
            // 6.存在，写入redis  
            stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, JSONUtil.toJsonStr(shop), RedisConstants.CACHE_SHOP_TTL, TimeUnit.MINUTES);  
        } catch (InterruptedException e) {  
            throw new RuntimeException(e);  
        } finally {  
            // 7.释放互斥锁  
            unLock(lockKey);  
        }  
  
        // 8.返回  
        return shop;  
    }  
  
    private Shop queryWithPassThrough(Long id) {  
        // 1.从redis查询商铺缓存  
        String shopJson = stringRedisTemplate.opsForValue().get(RedisConstants.CACHE_SHOP_KEY + id);  
  
        // 2.判断是否存在  
        if (StrUtil.isNotBlank(shopJson)) {  
            // 3.存在，直接返回  
            return JSONUtil.toBean(shopJson, Shop.class);  
        }  
  
        // 判断命中的是否是空值  
        if (shopJson != null) {  
            // 返回一个错误信息  
            return null;  
        }  
  
        // 4.不存在，根据id查询数据库  
        Shop shop = getById(id);  
  
        // 5.不存在，返回错误  
        if (shop == null) {  
            // 将空值写入redis  
            stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, "", RedisConstants.CACHE_NULL_TTL, TimeUnit.MINUTES);  
            // 返回错误信息  
            return null;  
        }  
  
        // 6.存在，写入redis  
        stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, JSONUtil.toJsonStr(shop), RedisConstants.CACHE_SHOP_TTL, TimeUnit.MINUTES);  
  
        // 7.返回  
        return shop;  
    }  
  
    private boolean tryLock(String key) {  
        Boolean flag = stringRedisTemplate.opsForValue().setIfAbsent(key, "1", 10, TimeUnit.SECONDS);  
        return BooleanUtil.isTrue(flag);  
    }  
  
    private void unLock(String key) {  
        stringRedisTemplate.delete(key);  
    }  
  
    @Override  
    @Transactional    public Result update(Shop shop) {  
        Long id = shop.getId();  
        if (id == null) {  
            return Result.fail("店铺id不能为空！");  
        }  
        // 1.更新数据库  
        updateById(shop);  
  
        // 2.删除缓存  
        stringRedisTemplate.delete(RedisConstants.CACHE_SHOP_KEY + id);  
  
        return Result.ok();  
    }  
}
```

#### 案例：基于逻辑过期方式解决缓存击穿问题

需求：修改根据id查询商铺的业务，基于逻辑过期方式来解决缓存击穿问题
![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241114194636.png)

ShopServiceImpl.java

```java
@Service  
public class ShopServiceImpl extends ServiceImpl<ShopMapper, Shop> implements IShopService {  
  
    @Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    @Override  
    public Result queryById(Long id) {  
        // 缓存穿透  
//        Shop shop = queryWithPassThrough(id);  
  
        // 互斥锁解决缓存击穿  
//        Shop shop = queryWithMutex(id);  
  
        // 逻辑过期解决缓存击穿  
        Shop shop = queryWithLogicalExpire(id);  
        if (shop == null) {  
            return Result.fail("店铺不存在！");  
        }  
        return Result.ok(shop);  
    }  
  
    private static final ExecutorService CACHE_REBUILD_EXECUTOR = Executors.newFixedThreadPool(10);  
  
    private Shop queryWithLogicalExpire(Long id) {  
        // 1.从redis查询商铺缓存  
        String shopJson = stringRedisTemplate.opsForValue().get(RedisConstants.CACHE_SHOP_KEY + id);  
  
        // 2.判断是否存在  
        if (StrUtil.isBlank(shopJson)) {  
            // 3.不存在，直接返回  
            return null;  
        }  
  
        // 4.命中，需要先把JSON反序列化为对象  
        RedisData redisData = JSONUtil.toBean(shopJson, RedisData.class);  
        JSONObject data = (JSONObject) redisData.getData();  
        Shop shop = JSONUtil.toBean(data, Shop.class);  
        LocalDateTime expireTime = redisData.getExpireTime();  
  
        // 5.判断是否过期  
        if (expireTime.isAfter(LocalDateTime.now())) {  
            // 5.1 未过期，直接返回店铺信息  
            return shop;  
        }  
  
        // 5.2已过期 需要缓存重建  
        // 6.缓存重建  
        // 6.1获取互斥锁  
        String lockKey = RedisConstants.LOCK_SHOP_KEY + id;  
  
        // 6.2判断是否获取互斥锁成功  
        boolean isLock = tryLock(lockKey);  
        if (isLock) {  
            // 6.3成功，开启独立线程实现缓存重建  
            CACHE_REBUILD_EXECUTOR.submit(() -> {  
                try {  
                    // 重建缓存  
                    this.saveShop2Redis(id, 20L);  
                } catch (Exception e) {  
                    throw new RuntimeException(e);  
                } finally {  
                    // 释放锁  
                    unLock(lockKey);  
                }  
            });  
        }  
  
        // 6.4返回过期的商铺信息  
        return shop;  
    }  
  
    private Shop queryWithMutex(Long id) {  
        // 1.从redis查询商铺缓存  
        String shopJson = stringRedisTemplate.opsForValue().get(RedisConstants.CACHE_SHOP_KEY + id);  
  
        // 2.判断是否存在  
        if (StrUtil.isNotBlank(shopJson)) {  
            // 3.存在，直接返回  
            return JSONUtil.toBean(shopJson, Shop.class);  
        }  
  
        // 判断命中的是否是空值  
        if (shopJson != null) {  
            // 返回一个错误信息  
            return null;  
        }  
  
        // 4.开始实现缓存重建  
        // 4.1获取互斥锁  
        String lockKey = RedisConstants.LOCK_SHOP_KEY + id;  
        Shop shop = null;  
        try {  
            boolean isLock = tryLock(lockKey);  
  
            // 4.2判断是否获取成功  
            if (!isLock) {  
                // 4.3失败，休眠并重试  
                Thread.sleep(50);  
                return queryWithMutex(id);  
            }  
  
            // 4.4成功，根据id查询数据库  
            shop = getById(id);  
  
            // 模拟重建延时  
            Thread.sleep(200);  
  
            // 5.不存在，返回错误  
            if (shop == null) {  
                // 将空值写入redis  
                stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, "", RedisConstants.CACHE_NULL_TTL, TimeUnit.MINUTES);  
                // 返回错误信息  
                return null;  
            }  
  
            // 6.存在，写入redis  
            stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, JSONUtil.toJsonStr(shop), RedisConstants.CACHE_SHOP_TTL, TimeUnit.MINUTES);  
        } catch (InterruptedException e) {  
            throw new RuntimeException(e);  
        } finally {  
            // 7.释放互斥锁  
            unLock(lockKey);  
        }  
  
        // 8.返回  
        return shop;  
    }  
  
    private Shop queryWithPassThrough(Long id) {  
        // 1.从redis查询商铺缓存  
        String shopJson = stringRedisTemplate.opsForValue().get(RedisConstants.CACHE_SHOP_KEY + id);  
  
        // 2.判断是否存在  
        if (StrUtil.isNotBlank(shopJson)) {  
            // 3.存在，直接返回  
            return JSONUtil.toBean(shopJson, Shop.class);  
        }  
  
        // 判断命中的是否是空值  
        if (shopJson != null) {  
            // 返回一个错误信息  
            return null;  
        }  
  
        // 4.不存在，根据id查询数据库  
        Shop shop = getById(id);  
  
        // 5.不存在，返回错误  
        if (shop == null) {  
            // 将空值写入redis  
            stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, "", RedisConstants.CACHE_NULL_TTL, TimeUnit.MINUTES);  
            // 返回错误信息  
            return null;  
        }  
  
        // 6.存在，写入redis  
        stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, JSONUtil.toJsonStr(shop), RedisConstants.CACHE_SHOP_TTL, TimeUnit.MINUTES);  
  
        // 7.返回  
        return shop;  
    }  
  
    private boolean tryLock(String key) {  
        Boolean flag = stringRedisTemplate.opsForValue().setIfAbsent(key, "1", 10, TimeUnit.SECONDS);  
        return BooleanUtil.isTrue(flag);  
    }  
  
    private void unLock(String key) {  
        stringRedisTemplate.delete(key);  
    }  
  
    public void saveShop2Redis(Long id, Long expireSeconds) throws InterruptedException {  
        // 1.查询店铺数据  
        Shop shop = getById(id);  
        Thread.sleep(200);  
  
        // 2.封装逻辑过期时间  
        RedisData redisData = new RedisData();  
        redisData.setData(shop);  
        redisData.setExpireTime(LocalDateTime.now().plusSeconds(expireSeconds));  
  
        // 3.写入redis  
        stringRedisTemplate.opsForValue().set(RedisConstants.CACHE_SHOP_KEY + id, JSONUtil.toJsonStr(redisData));  
    }  
  
    @Override  
    @Transactional    public Result update(Shop shop) {  
        Long id = shop.getId();  
        if (id == null) {  
            return Result.fail("店铺id不能为空！");  
        }  
        // 1.更新数据库  
        updateById(shop);  
  
        // 2.删除缓存  
        stringRedisTemplate.delete(RedisConstants.CACHE_SHOP_KEY + id);  
  
        return Result.ok();  
    }  
}
```

## 缓存工具封装

基于StringRedisTemplate封装一个缓存工具类，满足下列需求：

- 方法1：将任意Java对象序列化为json并存储在string类型的key中，并且可以设置TTL过期时间
- 方法2：将任意Java对象序列化为json并存储在string类型的key中，并且可以设置逻辑过期时间，用于处理缓存击穿问题
- 方法3：根据指定的key查询缓存，并反序列化为指定类型，利用缓存空值的方式解决缓存穿透问题
- 方法4：根据指定的key查询缓存，并反序列化为指定类型，需要利用逻辑过期解决缓存击穿问题

CacheClient.java

```java
@Component  
@Slf4j  
public class CacheClient {  
  
    private final StringRedisTemplate stringRedisTemplate;  
  
    public CacheClient(StringRedisTemplate stringRedisTemplate) {  
        this.stringRedisTemplate = stringRedisTemplate;  
    }  
  
    public void set(String key, Object value, Long time, TimeUnit unit) {  
        stringRedisTemplate.opsForValue().set(key, JSONUtil.toJsonStr(value), time, unit);  
    }  
  
    public void setWithLogicalExpire(String key, Object value, Long time, TimeUnit unit) {  
        // 设置逻辑过期  
        RedisData redisData = new RedisData();  
        redisData.setData(value);  
        redisData.setExpireTime(LocalDateTime.now().plusSeconds(unit.toSeconds(time)));  
  
        // 写入redis  
        stringRedisTemplate.opsForValue().set(key, JSONUtil.toJsonStr(redisData));  
    }  
  
    public  <R, ID> R queryWithPassThrough(String keyPrefix, ID id, Class<R> type, Function<ID, R> dbFallback, Long time, TimeUnit unit) {  
        String key = keyPrefix + id;  
        // 1.从redis查询商铺缓存  
        String json = stringRedisTemplate.opsForValue().get(key);  
  
        // 2.判断是否存在  
        if (StrUtil.isNotBlank(json)) {  
            // 3.存在，直接返回  
            return JSONUtil.toBean(json, type);  
        }  
  
        // 判断命中的是否是空值  
        if (json != null) {  
            // 返回一个错误信息  
            return null;  
        }  
  
        // 4.不存在，根据id查询数据库  
        R r = dbFallback.apply(id);  
  
        // 5.不存在，返回错误  
        if (r == null) {  
            // 将空值写入redis  
            stringRedisTemplate.opsForValue().set(key, "", RedisConstants.CACHE_NULL_TTL, TimeUnit.MINUTES);  
            // 返回错误信息  
            return null;  
        }  
  
        // 6.存在，写入redis  
        this.set(key, r, time, unit);  
  
        // 7.返回  
        return r;  
    }  
  
    private static final ExecutorService CACHE_REBUILD_EXECUTOR = Executors.newFixedThreadPool(10);  
  
    public <R, ID> R queryWithLogicalExpire(String keyPrefix, ID id, Class<R> type, Function<ID, R> dbFallback, Long time, TimeUnit unit) {  
        String key = keyPrefix + id;  
        // 1.从redis查询商铺缓存  
        String json = stringRedisTemplate.opsForValue().get(key);  
  
        // 2.判断是否存在  
        if (StrUtil.isBlank(json)) {  
            // 3.不存在，直接返回  
            return null;  
        }  
  
        // 4.命中，需要先把JSON反序列化为对象  
        RedisData redisData = JSONUtil.toBean(json, RedisData.class);  
        R r = JSONUtil.toBean((JSONObject) redisData.getData(), type);  
        LocalDateTime expireTime = redisData.getExpireTime();  
  
        // 5.判断是否过期  
        if (expireTime.isAfter(LocalDateTime.now())) {  
            // 5.1 未过期，直接返回店铺信息  
            return r;  
        }  
  
        // 5.2已过期 需要缓存重建  
        // 6.缓存重建  
        // 6.1获取互斥锁  
        String lockKey = RedisConstants.LOCK_SHOP_KEY + id;  
  
        // 6.2判断是否获取互斥锁成功  
        boolean isLock = tryLock(lockKey);  
        if (isLock) {  
            // 6.3成功，开启独立线程实现缓存重建  
            CACHE_REBUILD_EXECUTOR.submit(() -> {  
                try {  
                    // 重建缓存  
                    // 查询数据库  
                    R r1 = dbFallback.apply(id);  
  
                    // 写入redis  
                    this.setWithLogicalExpire(key, r1, time, unit);  
                } catch (Exception e) {  
                    throw new RuntimeException(e);  
                } finally {  
                    // 释放锁  
                    unLock(lockKey);  
                }  
            });  
        }  
  
        // 6.4返回过期的商铺信息  
        return r;  
    }  
  
    private boolean tryLock(String key) {  
        Boolean flag = stringRedisTemplate.opsForValue().setIfAbsent(key, "1", 10, TimeUnit.SECONDS);  
        return BooleanUtil.isTrue(flag);  
    }  
  
    private void unLock(String key) {  
        stringRedisTemplate.delete(key);  
    }  
  
}
```

ShopServiceImpl.java

```java
@Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    @Resource  
    private CacheClient cacheClient;  
  
    @Override  
    public Result queryById(Long id) {  
        // 缓存穿透  
//        Shop shop = cacheClient.queryWithPassThrough(RedisConstants.CACHE_SHOP_KEY, id, Shop.class, this::getById,  
//                RedisConstants.CACHE_SHOP_TTL, TimeUnit.MINUTES);  
  
        // 互斥锁解决缓存击穿  
//        Shop shop = queryWithMutex(id);  
  
        // 逻辑过期解决缓存击穿  
        Shop shop = cacheClient.queryWithLogicalExpire(RedisConstants.CACHE_SHOP_KEY, id, Shop.class, this::getById,  
                20L, TimeUnit.SECONDS);  
        if (shop == null) {  
            return Result.fail("店铺不存在！");  
        }  
        return Result.ok(shop);  
    }
```

HmDianPingApplicationTests.java

```java
@SpringBootTest  
class HmDianPingApplicationTests {  
  
    @Resource  
    private ShopServiceImpl shopService;  
  
    @Resource  
    private CacheClient cacheClient;  
  
    @Test  
    void testSaveShop() throws InterruptedException {  
        Shop shop = shopService.getById(1L);  
  
        cacheClient.setWithLogicalExpire(RedisConstants.CACHE_SHOP_KEY + 1L, shop, 10L, TimeUnit.SECONDS);  
    }  
}
```

# 优惠券秒杀

## 全局唯一ID

每个店铺都可以发布优惠券：
![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241114225213.png)

当用户抢购时，就会生成订单并保存到tb_voucher_order这张表中，而订单表如果使用数据库自增ID就存在一些问题：

- id的规律性太明显
- 受单表数据量的限制

### 全局ID生成器

全局ID生成器，是一种在分布式系统下用来生成全局唯一ID的工具，一般要满足下列特性：
![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241114225830.png)

为了增加ID的安全性，我们可以不直接使用Redis自增的数值，而是拼接一些其它信息：
![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241114230442.png)

ID的组成部分：

- 符号位：1bit，永远为0
- 时间戳：31bit，以秒为单位，可以使用69年
- 序列号：32bit，秒内的计数器，支持每秒产生2^32个不同ID

全局唯一ID生成策略：

- UUID
- Redis自增
- snowflake算法
- 数据库自增
  Redis自增ID策略：
- 每天一个key，方便统计订单量
- ID构造是 时间戳 + 计数器

RedisIdWorker.java

```java
@Component  
public class RedisIdWorker {  
  
    /**  
     * 开始时间戳  
     */  
    private static final long BEGIN_TIMESTAMP = 1704067200;  
  
    /**  
     * 序列号的位数  
     */  
    private static final long COUNT_BITS = 32;  
  
    @Resource  
    private StringRedisTemplate stringRedisTemplate;  
  
    public long nextId(String keyPrefix) {  
        // 1.生成时间戳  
        LocalDateTime now = LocalDateTime.now();  
        long nowSecond = now.toEpochSecond(ZoneOffset.UTC);  
        long timestamp = nowSecond - BEGIN_TIMESTAMP;  
  
        // 2.生成序列号  
        // 2.1获取当前的日期，精确到天  
        String date = now.format(DateTimeFormatter.ofPattern("yyyy:MM:dd"));  
  
        // 2.2自增长  
        long count = stringRedisTemplate.opsForValue().increment("icr:" + keyPrefix + ":" + date);  
  
        // 3.拼接并返回  
        return timestamp << COUNT_BITS | count;  
    }  
  
}
```

## 实现优惠券秒杀下单

每个店铺都可以发布优惠券，分为平价券和特价券。平价券可以任意购买，而特价券需要秒杀抢购：
![image.png](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/20241118200225.png)

表关系如下：

- tb_voucher：优惠券的基本信息，优惠金额、使用规则等
- tb_seckill_voucher：优惠券的库存、开始抢购时间，结束抢购时间。特价优惠券才需要填写这些信息

在VoucherController中提供了一个接口，可以添加秒杀优惠券：

![image-20241118203112841](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118203112841.png)



用户可以在店铺页面中抢购这些优惠券：

![image-20241118203140915](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118203140915.png)

下单时需要判断两点：

- 秒杀是否开始或结束，如果尚未开始或已经结束则无法下单

- 库存是否充足，不足则无法下单

![image-20241118203518776](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118203518776.png)

```java
@RestController
@RequestMapping("/voucher-order")
public class VoucherOrderController {

    @Resource
    private IVoucherOrderService voucherOrderService;

    @PostMapping("seckill/{id}")
    public Result seckillVoucher(@PathVariable("id") Long voucherId) {
        return voucherOrderService.seckillVoucher(voucherId);
    }
}
```

```java
public interface IVoucherOrderService extends IService<VoucherOrder> {

    Result seckillVoucher(Long voucherId);
}
```

```java
@Service
public class VoucherOrderServiceImpl extends ServiceImpl<VoucherOrderMapper, VoucherOrder> implements IVoucherOrderService {

    @Resource
    private ISeckillVoucherService seckillVoucherService;

    @Resource
    private RedisIdWorker redisIdWorker;

    @Override
    @Transactional
    public Result seckillVoucher(Long voucherId) {
        // 1.查询优惠券
        SeckillVoucher voucher = seckillVoucherService.getById(voucherId);

        // 2.判断秒杀是否开始
        if (voucher.getBeginTime().isAfter(LocalDateTime.now())) {
            // 尚未开始
            return Result.fail("秒杀尚未开始！");
        }

        // 3.判断秒杀是否已经结束
        if (voucher.getEndTime().isBefore(LocalDateTime.now())) {
            // 秒杀已经结束
            return Result.fail("秒杀已经结束！");
        }

        // 4.判断库存是否充足
        if (voucher.getStock() < 1) {
            // 库存不足
            return Result.fail("库存不足！");
        }

        // 5.扣减库存
        boolean success = seckillVoucherService.update().setSql("stock = stock - 1").eq("voucher_id", voucherId).update();
        if(!success) {
            // 扣减失败
            return Result.fail("库存不足！");
        }

        // 6.创建订单
        VoucherOrder voucherOrder = new VoucherOrder();

        // 6.1 订单id
        long orderId = redisIdWorker.nextId("order");
        voucherOrder.setId(orderId);

        // 6.2 用户id
        Long userId = UserHolder.getUser().getId();
        voucherOrder.setUserId(userId);

        // 6.3 代金券id
        voucherOrder.setVoucherId(voucherId);
        save(voucherOrder);

        // 7.返回订单id
        return Result.ok(orderId);
    }
}
```

## 超卖问题

![image-20241118210620126](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118210620126.png)

![image-20241118210836831](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118210836831.png)

超卖问题是典型的多线程安全问题，针对这一问题的常见解决方案就是加锁：

![image-20241118210920126](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118210920126.png)



### 乐观锁

乐观锁的关键是判断之前查询得到的数据是否有被修改过，常见的方式有两种：

#### 版本号法

![image-20241118211800538](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118211800538.png)

#### CAS法

![image-20241118212044684](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118212044684.png)

超卖这样的线程安全问题，解决方案有哪些？

1. 悲观锁：添加同步锁，让线程串行执行
   - 优点：简单粗暴
   - 缺点：性能一般

2. 乐观锁：不加锁，在更新时判断是否有其它线程在修改
   - 优点：性能好
   - 缺点：存在成功率低的问题

## 一人一单

需求：修改秒杀业务，要求同一个优惠券，一个用户只能下一单

![image-20241118213446900](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118213446900.png)

![image-20241118213510168](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118213510168.png)

### 一人一单的并发安全问题

通过加锁可以解决在单机情况下的一人一单安全问题，但是在集群模式下就不行了。

1. 我们将服务启动两份，端口分别为8081和8082：

2. 然后修改nginx的conf目录下的nginx.conf文件，配置反向代理和负载均衡：

现在，用户请求会在这两个节点上负载均衡，再次测试下是否存在线程安全问题。

![image-20241118215015284](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118215015284.png)

![image-20241118215044387](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118215044387.png)

![image-20241118215121699](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118215121699.png)

## 分布式锁

![image-20241118225912579](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118225912579.png)

![image-20241118225953459](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118225953459.png)

### 什么是分布式锁

**分布式锁：**满足分布式系统或集群模式下多进程可见并且互斥的锁。

![image-20241118230415911](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118230415911.png)

### 分布式锁的实现

分布式锁的核心是实现多进程之间互斥，而满足这一点的方式有很多，常见的有三种：

![image-20241118230503987](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241118230503987.png)

### 基于Redis的分布式锁

实现分布式锁时需要实现的两个基本方法：

- 获取锁：

  - 互斥：确保只能有一个线程获取锁

    ![image-20241119184013035](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119184013035.png)

  - 非阻塞：尝试一次，成功返回true，失败返回false

    ![image-20241119184113808](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119184113808.png)

- 释放锁：

  - 手动释放

  - 超时释放：获取锁时添加一个超时时间

    ![image-20241119184136265](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119184136265.png)

![image-20241119184226790](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119184226790.png)

#### 基于Redis实现分布式锁初级版本

需求：定义一个类，实现下面接口，利用Redis实现分布式锁功能。

```java
public interface ILock {

    /**
     * 尝试获取锁
     * @param timeoutSec 锁持有的超时时间，过期后自动释放
     * @return true代表获取锁成功；false代表获取锁失败
     */
    boolean tryLock(long timeoutSec);

    /**
     * 释放锁
     */
    void unlock();
}
```

```java
public class SimpleRedisLock implements ILock{

    private String name;
    private StringRedisTemplate stringRedisTemplate;

    public SimpleRedisLock(String name, StringRedisTemplate stringRedisTemplate) {
        this.name = name;
        this.stringRedisTemplate = stringRedisTemplate;
    }

    private static final String KEY_PREFIX = "lock:";

    @Override
    public boolean tryLock(long timeoutSec) {
        // 获取线程标识
        long threadId = Thread.currentThread().getId();

        // 获取锁
        Boolean success = stringRedisTemplate.opsForValue().setIfAbsent(KEY_PREFIX + name, threadId + "", timeoutSec, TimeUnit.SECONDS);
        return Boolean.TRUE.equals(success);
    }

    @Override
    public void unlock() {
        // 释放锁
        stringRedisTemplate.delete(KEY_PREFIX + name);
    }

}
```

```java
@Service
public class VoucherOrderServiceImpl extends ServiceImpl<VoucherOrderMapper, VoucherOrder> implements IVoucherOrderService {

    @Resource
    private ISeckillVoucherService seckillVoucherService;

    @Resource
    private RedisIdWorker redisIdWorker;

    @Resource
    private StringRedisTemplate stringRedisTemplate;

    @Override
    public Result seckillVoucher(Long voucherId) {
        // 1.查询优惠券
        SeckillVoucher voucher = seckillVoucherService.getById(voucherId);

        // 2.判断秒杀是否开始
        if (voucher.getBeginTime().isAfter(LocalDateTime.now())) {
            // 尚未开始
            return Result.fail("秒杀尚未开始！");
        }

        // 3.判断秒杀是否已经结束
        if (voucher.getEndTime().isBefore(LocalDateTime.now())) {
            // 秒杀已经结束
            return Result.fail("秒杀已经结束！");
        }

        // 4.判断库存是否充足
        if (voucher.getStock() < 1) {
            // 库存不足
            return Result.fail("库存不足！");
        }

        Long userId = UserHolder.getUser().getId();
        // 创建锁对象
        SimpleRedisLock lock = new SimpleRedisLock("order:" + userId, stringRedisTemplate);

        // 获取锁
        boolean isLock = lock.tryLock(5L);

        // 判断是否获取锁成功
        if (!isLock) {
            // 获取锁失败，返回错误信息或重试
            return Result.fail("不允许重复下单！");
        }
        try {
            // 获取代理对象（事务）
            IVoucherOrderService proxy = (IVoucherOrderService) AopContext.currentProxy();
            return proxy.createVoucherOrder(voucherId, voucher);
        } finally {
            // 释放锁
            lock.unlock();
        }
    }
}
```



![image-20241119192657573](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119192657573.png)

![image-20241119192730169](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119192730169.png)

#### 改进Redis的分布式锁

需求：修改之前的分布式锁实现，满足：

1. 在获取锁时存入线程标示（可以用UUID表示）

2. 在释放锁时先获取锁中的线程标示，判断是否与当前线程标示一致
   - 如果一致则释放锁
   - 如果不一致则不释放锁

```java
public class SimpleRedisLock implements ILock{

    private String name;
    private StringRedisTemplate stringRedisTemplate;

    public SimpleRedisLock(String name, StringRedisTemplate stringRedisTemplate) {
        this.name = name;
        this.stringRedisTemplate = stringRedisTemplate;
    }

    private static final String KEY_PREFIX = "lock:";
    private static final String ID_PREFIX = UUID.randomUUID().toString(true);

    @Override
    public boolean tryLock(long timeoutSec) {
        // 获取线程标识
        String threadId = ID_PREFIX + Thread.currentThread().getId();

        // 获取锁
        Boolean success = stringRedisTemplate.opsForValue().setIfAbsent(KEY_PREFIX + name, threadId, timeoutSec, TimeUnit.SECONDS);
        return Boolean.TRUE.equals(success);
    }

    @Override
    public void unlock() {
        // 获取线程标识
        String threadId = ID_PREFIX + Thread.currentThread().getId();
        // 获取锁中的标识
        String id = stringRedisTemplate.opsForValue().get(KEY_PREFIX + name);
        // 判断标识是否一致
        if (threadId.equals(id)) {
            // 释放锁
            stringRedisTemplate.delete(KEY_PREFIX + name);
        }
    }

}
```

![image-20241119200321003](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119200321003.png)

#### Redis的Lua脚本

Redis提供了Lua脚本功能，在一个脚本中编写多条Redis命令，确保多条命令执行时的原子性。Lua是一种编程语言，它的基本语法大家可以参考网站：https://www.runoob.com/lua/lua-tutorial.html

这里重点介绍Redis提供的调用函数，语法如下：

![image-20241119202829354](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119202829354.png)

例如，我们要执行set name jack，则脚本是这样：

![image-20241119202840801](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119202840801.png)

例如，我们要先执行set name Rose，再执行get name，则脚本如下：

![image-20241119202851830](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119202851830.png)

写好脚本以后，需要用Redis命令来调用脚本，调用脚本的常见命令如下：

![](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119202905122.png)

例如，我们要执行 redis.call('set', 'name', 'jack') 这个脚本，语法如下：

![image-20241119203244327](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119203244327.png)

如果脚本中的key、value不想写死，可以作为参数传递。key类型参数会放入KEYS数组，其它参数会放入ARGV数组，在脚本中可以从KEYS和ARGV数组获取这些参数：

![image-20241119202954838](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119202954838.png)

释放锁的业务流程是这样的：

1. 获取锁中的线程标示
2. 判断是否与指定的标示（当前线程标示）一致
3. 如果一致则释放锁（删除）
4. 如果不一致则什么都不做

如果用Lua脚本来表示则是这样的：

![image-20241119203750535](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119203750535.png)

#### 再次改进Redis的分布式锁

需求：基于Lua脚本实现分布式锁的释放锁逻辑

提示：RedisTemplate调用Lua脚本的API如下：

![image-20241119205012145](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241119205012145.png)

```lua
-- 比较线程标识与锁中的标识是否一致
if(redis.call('get', KEYS[1]) == ARGV[1]) then
    -- 释放锁 del key
    return redis.call('del', KEYS[1])
end
return 0
```

```java
public class SimpleRedisLock implements ILock{

    private String name;
    private StringRedisTemplate stringRedisTemplate;

    public SimpleRedisLock(String name, StringRedisTemplate stringRedisTemplate) {
        this.name = name;
        this.stringRedisTemplate = stringRedisTemplate;
    }

    private static final String KEY_PREFIX = "lock:";
    private static final String ID_PREFIX = UUID.randomUUID().toString(true) + "-";
    private static final DefaultRedisScript<Long> UNLOCK_SCRIPT;
    static {
        UNLOCK_SCRIPT = new DefaultRedisScript<>();
        UNLOCK_SCRIPT.setLocation(new ClassPathResource("unlock.lua"));
        UNLOCK_SCRIPT.setResultType(Long.class);
    }

    @Override
    public boolean tryLock(long timeoutSec) {
        // 获取线程标识
        String threadId = ID_PREFIX + Thread.currentThread().getId();

        // 获取锁
        Boolean success = stringRedisTemplate.opsForValue().setIfAbsent(KEY_PREFIX + name, threadId, timeoutSec, TimeUnit.SECONDS);
        return Boolean.TRUE.equals(success);
    }

    @Override
    public void unlock() {
        // 调用Lua脚本
        stringRedisTemplate.execute(UNLOCK_SCRIPT, Collections.singletonList(KEY_PREFIX + name), ID_PREFIX + Thread.currentThread().getId());
    }
}
```

基于Redis的分布式锁实现思路：

- 利用set nx ex获取锁，并设置过期时间，保存线程标示

- 释放锁时先判断线程标示是否与自己一致，一致则删除锁

特性：

- 利用set nx满足互斥性

- 利用set ex保证故障时锁依然能释放，避免死锁，提高安全性

- 利用Redis集群保证高可用和高并发特性

#### 基于Redis的分布式锁优化

基于setnx实现的分布式锁存在下面的问题：

![image-20241120185437715](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120185437715.png)

### Redisson

Redisson是一个在Redis的基础上实现的Java驻内存数据网格（In-Memory Data Grid）。它不仅提供了一系列的分布式的Java常用对象，还提供了许多分布式服务，其中就包含了各种分布式锁的实现。

![image-20241120190314580](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120190314580.png)

官网地址： [https://redisson.org](https://redisson.org/)

GitHub地址： https://github.com/redisson/redisson

#### Redisson入门

1. 引入依赖：

   ```xml
   <dependency>
       <groupId>org.redisson</groupId>
       <artifactId>redisson</artifactId>
       <version>3.13.6</version>
   </dependency>
   ```

2. 配置Redisson客户端

   ![image-20241120190727178](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120190727178.png)

3. 使用Redisson的分布式锁

   ![image-20241120190752522](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120190752522.png)

   把之前的锁对象改成Redisson的锁

   ```java
           // 创建锁对象
   //        SimpleRedisLock lock = new SimpleRedisLock("order:" + userId, stringRedisTemplate);
           RLock lock = redissonClient.getLock("lock:order:" + userId);
   
           // 获取锁
           boolean isLock = lock.tryLock();
   ```

#### Redisson可重入锁原理

![image-20241120192648879](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120192648879.png)

![image-20241120193121320](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120193121320.png)

获取锁的Lua脚本：

![image-20241120193905902](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120193905902.png)

释放锁的Lua脚本：

![image-20241120195224362](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120195224362.png)

#### Redisson分布式锁原理

![image-20241120201824413](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120201824413.png)

Redisson分布式锁原理：

- **可重入**：利用hash结构记录线程id和重入次数
- **可重试**：利用信号量和PubSub功能实现等待、唤醒，获取锁失败的重试机制
- **超时续约**：利用watchDog，每隔一段时间（releaseTime / 3），重置超时时间

#### Redisson分布式锁主从一致性问题

![image-20241120202203523](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120202203523.png)

![image-20241120202215978](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120202215978.png)

![image-20241120202224203](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120202224203.png)

![image-20241120202232461](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241120202232461.png)

**不可重入Redis分布式锁：**

- 原理：利用setnx的互斥性；利用ex避免死锁；释放锁时判断线程标示
- 缺陷：不可重入、无法重试、锁超时失效

**可重入的Redis分布式锁：**

- 原理：利用hash结构，记录线程标示和重入次数；利用watchDog延续锁时间；利用信号量控制锁重试等待
- 缺陷：redis宕机引起锁失效问题

**Redisson的multiLock：**

- 原理：多个独立的Redis节点，必须在所有节点都获取重入锁，才算获取锁成功
- 缺陷：运维成本高、实现复杂

## Redis优化秒杀

![image-20241125192601303](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241125192601303.png)

![image-20241125192620900](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241125192620900.png)

![image-20241125192638678](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241125192638678.png)

![image-20241125192736507](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241125192736507.png)

### 案例：改进秒杀业务，提高并发性能

需求：

①新增秒杀优惠券的同时，将优惠券信息保存到Redis中

②基于Lua脚本，判断秒杀库存、一人一单，决定用户是否抢购成功

③如果抢购成功，将优惠券id和用户id封装后存入阻塞队列

④开启线程任务，不断从阻塞队列中获取信息，实现异步下单功能

```java
@Service
public class VoucherServiceImpl extends ServiceImpl<VoucherMapper, Voucher> implements IVoucherService {

    @Resource
    private ISeckillVoucherService seckillVoucherService;

    @Resource
    private StringRedisTemplate stringRedisTemplate;

    @Override
    public Result queryVoucherOfShop(Long shopId) {
        // 查询优惠券信息
        List<Voucher> vouchers = getBaseMapper().queryVoucherOfShop(shopId);
        // 返回结果
        return Result.ok(vouchers);
    }

    @Override
    @Transactional
    public void addSeckillVoucher(Voucher voucher) {
        // 保存优惠券
        save(voucher);
        // 保存秒杀信息
        SeckillVoucher seckillVoucher = new SeckillVoucher();
        seckillVoucher.setVoucherId(voucher.getId());
        seckillVoucher.setStock(voucher.getStock());
        seckillVoucher.setBeginTime(voucher.getBeginTime());
        seckillVoucher.setEndTime(voucher.getEndTime());
        seckillVoucherService.save(seckillVoucher);

        // 保存秒杀库存到redis
        stringRedisTemplate.opsForValue().set(RedisConstants.SECKILL_STOCK_KEY + voucher.getId(), voucher.getStock().toString());
    }
}
```

```java
@Slf4j
@Service
public class VoucherOrderServiceImpl extends ServiceImpl<VoucherOrderMapper, VoucherOrder> implements IVoucherOrderService {

    @Resource
    private ISeckillVoucherService seckillVoucherService;

    @Resource
    private RedisIdWorker redisIdWorker;

    @Resource
    private StringRedisTemplate stringRedisTemplate;

    @Resource
    private RedissonClient redissonClient;

    private static final DefaultRedisScript<Long> SECKILL_SCRIPT;
    static {
        SECKILL_SCRIPT = new DefaultRedisScript<>();
        SECKILL_SCRIPT.setLocation(new ClassPathResource("seckill.lua"));
        SECKILL_SCRIPT.setResultType(Long.class);
    }

    private BlockingQueue<VoucherOrder> orderTasks = new ArrayBlockingQueue<>(1024 * 1024);
    private static final ExecutorService SECKILL_ORDER_EXECUTOR = Executors.newSingleThreadExecutor();

    @PostConstruct
    private void init() {
        SECKILL_ORDER_EXECUTOR.submit(new VoucherOrderHandler());
    }

    private class VoucherOrderHandler implements Runnable {

        @Override
        public void run() {
            while (true) {
                try {
                    // 1.获取队列中的订单信息
                    VoucherOrder voucherOrder = orderTasks.take();
                    // 2.创建订单
                    handleVoucherOrder(voucherOrder);
                } catch (InterruptedException e) {
                    log.error("处理订单异常", e);
                }
            }
        }
    }

    private void handleVoucherOrder(VoucherOrder voucherOrder) {
        // 1.获取用户
        Long userId = voucherOrder.getUserId();

        // 2.创建锁对象
        RLock lock = redissonClient.getLock("lock:order:" + userId);

        // 3.获取锁
        boolean isLock = lock.tryLock();

        // 4.判断是否获取锁成功
        if (!isLock) {
            // 获取锁失败，返回错误信息或重试
            log.error("不允许重复下单");
            return;
        }
        try {
            proxy.createVoucherOrder(voucherOrder);
        } finally {
            // 释放锁
            lock.unlock();
        }
    }

    private IVoucherOrderService proxy;
    @Override
    public Result seckillVoucher(Long voucherId) {
        // 获取用户
        Long userId = UserHolder.getUser().getId();
        // 1.执行lua脚本
        Long result = stringRedisTemplate.execute(
                SECKILL_SCRIPT,
                Collections.emptyList(),
                voucherId.toString(), userId.toString()
        );

        // 2.判断结果是否为0
        int r = result.intValue();
        if (r != 0) {
            // 2.1 不为0，代表没有购买资格
            return Result.fail(r == 1 ? "库存不足" : "不能重复下单");
        }

        // 2.2 为0，有购买资格，把下单信息保存到阻塞队列
        VoucherOrder voucherOrder = new VoucherOrder();

        // 2.3 订单id
        long orderId = redisIdWorker.nextId("order");
        voucherOrder.setId(orderId);

        // 2.4 用户id
        voucherOrder.setUserId(userId);

        // 2.5 代金券id
        voucherOrder.setVoucherId(voucherId);
        save(voucherOrder);

        // 2.6 创建阻塞队列
        orderTasks.add(voucherOrder);

        // 3.获取代理对象
        proxy = (IVoucherOrderService) AopContext.currentProxy();

        // 4.返回订单id
        return Result.ok(orderId);
    }

    /*@Override
    public Result seckillVoucher(Long voucherId) {
        // 1.查询优惠券
        SeckillVoucher voucher = seckillVoucherService.getById(voucherId);

        // 2.判断秒杀是否开始
        if (voucher.getBeginTime().isAfter(LocalDateTime.now())) {
            // 尚未开始
            return Result.fail("秒杀尚未开始！");
        }

        // 3.判断秒杀是否已经结束
        if (voucher.getEndTime().isBefore(LocalDateTime.now())) {
            // 秒杀已经结束
            return Result.fail("秒杀已经结束！");
        }

        // 4.判断库存是否充足
        if (voucher.getStock() < 1) {
            // 库存不足
            return Result.fail("库存不足！");
        }

        Long userId = UserHolder.getUser().getId();
        // 创建锁对象
//        SimpleRedisLock lock = new SimpleRedisLock("order:" + userId, stringRedisTemplate);
        RLock lock = redissonClient.getLock("lock:order:" + userId);

        // 获取锁
        boolean isLock = lock.tryLock();

        // 判断是否获取锁成功
        if (!isLock) {
            // 获取锁失败，返回错误信息或重试
            return Result.fail("不允许重复下单！");
        }
        try {
            // 获取代理对象（事务）
            IVoucherOrderService proxy = (IVoucherOrderService) AopContext.currentProxy();
            return proxy.createVoucherOrder(voucherId, voucher);
        } finally {
            // 释放锁
            lock.unlock();
        }
    }*/

    @Transactional
    public void createVoucherOrder(VoucherOrder voucherOrder) {
        // 5.一人一单
        Long userId = voucherOrder.getUserId();

        // 5.1 查询订单
        int count = query().eq("user_id", userId).eq("voucher_id", voucherOrder.getVoucherId()).count();

        // 5.2 判断是否存在
        if (count > 0) {
            // 用户已经购买过了
            log.error("用户已经购买过一次");
            return;
        }

        // 6.扣减库存
        boolean success = seckillVoucherService.update()
                .setSql("stock = stock - 1") // set stock = stock - 1
                .eq("voucher_id", voucherOrder)
                .gt("stock", 0) // where id = ? and stock > 0
                .update();
        if (!success) {
            // 扣减失败
            log.error("库存不足！");
            return;
        }

        save(voucherOrder);
    }
}
```

```java
public interface IVoucherOrderService extends IService<VoucherOrder> {

    Result seckillVoucher(Long voucherId);

    void createVoucherOrder(VoucherOrder voucherOrder);
}
```

```lua
-- 1.参数列表
-- 1.1 优惠券id
local voucherId = ARGV[1]
-- 1.2 用户id
local userId = ARGV[2]

-- 2.数据key
-- 2.1 库存key
local stockKey = 'seckill:stock:' .. voucherId
-- 2.2 订单key
local orderKey = 'seckill:order:' .. voucherId

-- 3.脚本业务
-- 3.1 判断库存是否充足 get stockKey
if(tonumber(redis.call('get', stockKey)) <= 0) then
    -- 3.2 库存不足，返回1
    return 1
end
-- 3.2 判断用户是否下单 SISMEMBER orderKey userId
if(redis.call('sismember'. orderKey, userId) == 1) then
    -- 3.3 存在，说明是重复下单
    return 2
end
-- 3.4 扣库存 incrby stockKey -1
redis.call('incrby', stockKey, -1)
-- 3.5 下单（保存用户） sadd orderKey userId
redis.call('sadd', orderKey, userId);
return 0;
```



秒杀业务的优化思路是什么？

①先利用Redis完成库存余量、一人一单判断，完成抢单业务

②再将下单业务放入阻塞队列，利用独立线程异步下单

基于阻塞队列的异步秒杀存在哪些问题？

- 内存限制问题
- 数据安全问题



## Redis消息队列实现异步秒杀

**消息队列**（**M**essage **Q**ueue），字面意思就是存放消息的队列。最简单的消息队列模型包括3个角色：

- 消息队列：存储和管理消息，也被称为消息代理（Message Broker）
- 生产者：发送消息到消息队列

- 消费者：从消息队列获取消息并处理消息

![image-20241126190134030](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126190134030.png)

Redis提供了三种不同的方式来实现消息队列：

- list结构：基于List结构模拟消息队列
- PubSub：基本的点对点消息模型
- Stream：比较完善的消息队列模型



### 基于List结构模拟消息队列

**消息队列**（**M**essage **Q**ueue），字面意思就是存放消息的队列。而Redis的list数据结构是一个**双向链表**，很容易模拟出队列效果。

队列是入口和出口不在一边，因此我们可以利用：LPUSH 结合 RPOP、或者 RPUSH 结合 LPOP来实现。

不过要注意的是，当队列中没有消息时RPOP或LPOP操作会返回null，并不像JVM的阻塞队列那样会阻塞并等待消息。因此这里应该使用**BRPOP**或者**BLPOP**来实现阻塞效果。

![image-20241126190932443](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126190932443.png)

基于List的消息队列有哪些优缺点？

优点：

- 利用Redis存储，不受限于JVM内存上限
- 基于Redis的持久化机制，数据安全性有保证
- 可以满足消息有序性

缺点：

- 无法避免消息丢失

- 只支持单消费者



### 基于PubSub的消息队列

**PubSub****（发布订阅）**是Redis2.0版本引入的消息传递模型。顾名思义，消费者可以订阅一个或多个channel，生产者向对应channel发送消息后，所有订阅者都能收到相关消息。

- SUBSCRIBE channel [channel] ：订阅一个或多个频道

- PUBLISH channel msg ：向一个频道发送消息

- PSUBSCRIBE pattern[pattern] ：订阅与pattern格式匹配的所有频道

![image-20241126192005395](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126192005395.png)

基于PubSub的消息队列有哪些优缺点？

优点：

- 采用发布订阅模型，支持多生产、多消费

缺点：

- 不支持数据持久化

- 无法避免消息丢失

- 消息堆积有上限，超出时数据丢失



### 基于Stream的消息队列

Stream 是 Redis 5.0 引入的一种新**数据类型**，可以实现一个功能非常完善的消息队列。

发送消息的命令：

![image-20241126192653439](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126192653439.png)

例如：

![image-20241126192711131](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126192711131.png)



#### 基于Stream的消息队列-XREAD

读取消息的方式之一：XREAD

![image-20241126193039936](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126193039936.png)

例如，使用XREAD读取第一个消息：

![image-20241126193059678](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126193059678.png)

XREAD阻塞方式，读取最新的消息：

![image-20241126193147291](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126193147291.png)

在业务开发中，我们可以循环的调用XREAD阻塞方式来查询最新消息，从而实现持续监听队列的效果，伪代码如下：

![image-20241126193159967](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126193159967.png)

**注意：**当我们指定起始ID为$时，代表读取最新的消息，如果我们处理一条消息的过程中，又有超过1条以上的消息到达队列，则下次获取时也只能获取到最新的一条，会出现**漏读消息**的问题。



STREAM类型消息队列的XREAD命令特点：

- 消息可回溯

- 一个消息可以被多个消费者读取

- 可以阻塞读取

- 有消息漏读的风险



#### 基于Stream的消息队列-消费者组

**消费者组（Consumer Group）**：将多个消费者划分到一个组中，监听同一个队列。具备下列特点：

![image-20241126200150341](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126200150341.png)

创建消费者组：

![image-20241126200727761](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126200727761.png)

- key：队列名称

- groupName：消费者组名称

- ID：起始ID标示，$代表队列中最后一个消息，0则代表队列中第一个消息

- MKSTREAM：队列不存在时自动创建队列

其它常见命令：

![image-20241126200750735](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126200750735.png)

从消费者组读取消息：

![image-20241126201431796](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126201431796.png)

- group：消费组名称
- consumer：消费者名称，如果消费者不存在，会自动创建一个消费者
- count：本次查询的最大数量
- BLOCK milliseconds：当没有消息时最长等待时间
- NOACK：无需手动ACK，获取到消息后自动确认
- STREAMS key：指定队列名称
- ID：获取消息的起始ID：
  - ">"：从下一个未消费的消息开始
  - 其它：根据指定id从pending-list中获取已消费但未确认的消息，例如0，是从pending-list中的第一个消息开始



消费者监听消息的基本思路：

![image-20241126203312342](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126203312342.png)



STREAM类型消息队列的XREADGROUP命令特点：

- 消息可回溯

- 可以多消费者争抢消息，加快消费速度

- 可以阻塞读取

- 没有消息漏读的风险
- 有消息确认机制，保证消息至少被消费一次



### Redis消息队列

![image-20241126203941027](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241126203941027.png)



### 案例：基于Redis的Stream结构作为消息队列，实现异步秒杀下单

需求：

①创建一个Stream类型的消息队列，名为stream.orders

②修改之前的秒杀下单Lua脚本，在认定有抢购资格后，直接向stream.orders中添加消息，内容包含voucherId、userId、orderId

③项目启动时，开启一个线程任务，尝试获取stream.orders中的消息，完成下单



seckill.lua

```lua
-- 1.参数列表
-- 1.1 优惠券id
local voucherId = ARGV[1]
-- 1.2 用户id
local userId = ARGV[2]
-- 1.3 订单id
local orderId = ARGV[3]

-- 2.数据key
-- 2.1 库存key
local stockKey = 'seckill:stock:' .. voucherId
-- 2.2 订单key
local orderKey = 'seckill:order:' .. voucherId

-- 3.脚本业务
-- 3.1 判断库存是否充足 get stockKey
if(tonumber(redis.call('get', stockKey)) <= 0) then
    -- 3.2 库存不足，返回1
    return 1
end
-- 3.2 判断用户是否下单 SISMEMBER orderKey userId
if(redis.call('sismember'. orderKey, userId) == 1) then
    -- 3.3 存在，说明是重复下单
    return 2
end
-- 3.4 扣库存 incrby stockKey -1
redis.call('incrby', stockKey, -1)
-- 3.5 下单（保存用户） sadd orderKey userId
redis.call('sadd', orderKey, userId);
-- 3.6 发送消息到队列中，xadd stream.orders * k1 v1 k2 v2 ...
redis.call('xadd', 'stream.orders', '*', 'userId', userId, 'voucherId', voucherId, 'id', orderId)
return 0;
```

VoucherOrderServiceImpl.java

```java
private class VoucherOrderHandler implements Runnable {

    @Override
    public void run() {
        while (true) {
            try {
                // 1.获取消息队列中的订单信息 xreadgroup group g1 c1 count 1 block 2000 streams streams.order >
                List<MapRecord<String, Object, Object>> list = stringRedisTemplate.opsForStream().read(
                        Consumer.from("g1", "c1"),
                        StreamReadOptions.empty().count(1).block(Duration.ofSeconds(2)),
                        StreamOffset.create("stream.orders", ReadOffset.lastConsumed())
                );
                // 2.判断消息获取是否成功
                if (list == null || list.isEmpty()) {
                    // 2.1 如果获取失败，说明没有消息，继续下一次循环
                    continue;
                }
                // 3.解析消息中的订单信息
                MapRecord<String, Object, Object> record = list.get(0);
                Map<Object, Object> values = record.getValue();
                VoucherOrder voucherOrder = BeanUtil.fillBeanWithMap(values, new VoucherOrder(), true);
                // 4.如果获取成功，可以下单
                handleVoucherOrder(voucherOrder);
                // 5.ack确认 sack stream.orders g1 id
                stringRedisTemplate.opsForStream().acknowledge("stream.orders", "g1", record.getId());
            } catch (Exception e) {
                log.error("处理订单异常", e);
                handlePendingList();
            }
        }
    }

    private void handlePendingList() {
        while (true) {
            try {
                // 1.获取pending-list中的订单信息 xreadgroup group g1 c1 count 1 streams streams.order 0
                List<MapRecord<String, Object, Object>> list = stringRedisTemplate.opsForStream().read(
                        Consumer.from("g1", "c1"),
                        StreamReadOptions.empty().count(1),
                        StreamOffset.create("stream.orders", ReadOffset.from("0"))
                );
                // 2.判断消息获取是否成功
                if (list == null || list.isEmpty()) {
                    // 2.1 如果获取失败，说明pending-list没有异常消息，结束循环
                    break;
                }
                // 3.解析消息中的订单信息
                MapRecord<String, Object, Object> record = list.get(0);
                Map<Object, Object> values = record.getValue();
                VoucherOrder voucherOrder = BeanUtil.fillBeanWithMap(values, new VoucherOrder(), true);
                // 4.如果获取成功，可以下单
                handleVoucherOrder(voucherOrder);
                // 5.ack确认 sack stream.orders g1 id
                stringRedisTemplate.opsForStream().acknowledge("stream.orders", "g1", record.getId());
            } catch (Exception e) {
                log.error("处理pending-list异常", e);
                try {
                    Thread.sleep(20);
                } catch (InterruptedException ex) {
                    ex.printStackTrace();
                }
            }
        }
    }

}
```



****



# 达人探店

## 发布探店笔记

探店笔记类似点评网站的评价，往往是图文结合。对应的表有两个：

- tb_blog：探店笔记表，包含笔记中的标题、文字、图片等

- tb_blog_comments：其他用户对探店笔记的评价

![image-20241127201238462](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241127201238462.png)

![image-20241127201244221](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241127201244221.png)

点击首页最下方菜单栏中的+按钮，即可发布探店图文：

![image-20241127201504934](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241127201504934.png)



****



### 案例：实现查看发布探店笔记的接口

需求：点击首页的探店笔记，会进入详情页面，实现该页面的查询接口：

![image-20241127202231407](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241127202231407.png)



BlogServiceImpl.java

```java
@Service
public class BlogServiceImpl extends ServiceImpl<BlogMapper, Blog> implements IBlogService {

    @Resource
    private IUserService userService;

    @Override
    public Result queryHotBlog(Integer current) {
        // 根据用户查询
        Page<Blog> page = query()
                .orderByDesc("liked")
                .page(new Page<>(current, SystemConstants.MAX_PAGE_SIZE));
        // 获取当前页的数据
        List<Blog> records = page.getRecords();
        // 查询用户
        records.forEach(this::queryBlogUser);
        return Result.ok(records);
    }

    @Override
    public Result queryBlogById(Long id) {
        // 1.查询blog
        Blog blog = getById(id);
        if (blog == null) {
            return Result.fail("笔记不存在！");
        }
        // 2.查询blog有关的用户
        queryBlogUser(blog);
        return Result.ok(blog);
    }

    private void queryBlogUser(Blog blog) {
        Long userId = blog.getUserId();
        User user = userService.getById(userId);
        blog.setName(user.getNickName());
        blog.setIcon(blog.getIcon());
    }
}
```



****



## 点赞

在首页的探店笔记排行榜和探店图文详情页面都有点赞的功能：

![image-20241127203801776](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241127203801776.png)



****



### 案例：完善点赞功能

需求：

- 同一个用户只能点赞一次，再次点击则取消点赞

- 如果当前用户已经点赞，则点赞按钮高亮显示（前端已实现，判断字段Blog类的isLike属性）



实现步骤：

①给Blog类中添加一个isLike字段，标示是否被当前用户点赞

②修改点赞功能，利用Redis的**set集合**判断是否点赞过，未点赞过则点赞数+1，已点赞过则点赞数-1

③修改根据id查询Blog的业务，判断当前登录用户是否点赞过，赋值给isLike字段

④修改分页查询Blog业务，判断当前登录用户是否点赞过，赋值给isLike字段



BlogController.java

```java
@RestController
@RequestMapping("/blog")
public class BlogController {

    @Resource
    private IBlogService blogService;

    @PostMapping
    public Result saveBlog(@RequestBody Blog blog) {
        // 获取登录用户
        UserDTO user = UserHolder.getUser();
        blog.setUserId(user.getId());
        // 保存探店博文
        blogService.save(blog);
        // 返回id
        return Result.ok(blog.getId());
    }

    @PutMapping("/like/{id}")
    public Result likeBlog(@PathVariable("id") Long id) {
        return blogService.likeBlog(id);
    }

    @GetMapping("/of/me")
    public Result queryMyBlog(@RequestParam(value = "current", defaultValue = "1") Integer current) {
        // 获取登录用户
        UserDTO user = UserHolder.getUser();
        // 根据用户查询
        Page<Blog> page = blogService.query()
                .eq("user_id", user.getId()).page(new Page<>(current, SystemConstants.MAX_PAGE_SIZE));
        // 获取当前页数据
        List<Blog> records = page.getRecords();
        return Result.ok(records);
    }

    @GetMapping("/hot")
    public Result queryHotBlog(@RequestParam(value = "current", defaultValue = "1") Integer current) {
        return blogService.queryHotBlog(current);
    }

    @GetMapping("/{id}")
    public Result queryBlogById(@PathVariable("id") Long id) {
        return blogService.queryBlogById(id);
    }
}
```

BlogServiceImpl.java

```java
@Service
public class BlogServiceImpl extends ServiceImpl<BlogMapper, Blog> implements IBlogService {

    @Resource
    private IUserService userService;

    @Resource
    private StringRedisTemplate stringRedisTemplate;

    @Override
    public Result queryHotBlog(Integer current) {
        // 根据用户查询
        Page<Blog> page = query()
                .orderByDesc("liked")
                .page(new Page<>(current, SystemConstants.MAX_PAGE_SIZE));
        // 获取当前页的数据
        List<Blog> records = page.getRecords();
        // 查询用户
        records.forEach(blog -> {
            this.queryBlogUser(blog);
            this.isBlockLiked(blog);
        });
        return Result.ok(records);
    }

    @Override
    public Result queryBlogById(Long id) {
        // 1.查询blog
        Blog blog = getById(id);
        if (blog == null) {
            return Result.fail("笔记不存在！");
        }
        // 2.查询blog有关的用户
        queryBlogUser(blog);
        // 3. 查询blog是否被点赞
        isBlockLiked(blog);
        return Result.ok(blog);
    }

    private void isBlockLiked(Blog blog) {
        // 1.获取登录用户
        Long userId = UserHolder.getUser().getId();
        // 2.判断当前登录用户是否已经点赞
        String key = "block:liked:" + blog.getId();
        Boolean isMember = stringRedisTemplate.opsForSet().isMember(key, userId.toString());
        blog.setIsLike(BooleanUtil.isTrue(isMember));
    }

    @Override
    public Result likeBlog(Long id) {
        // 1.获取登录用户
        Long userId = UserHolder.getUser().getId();
        // 2.判断当前登录用户是否已经点赞
        String key = "block:liked:" + id;
        Boolean isMember = stringRedisTemplate.opsForSet().isMember(key, userId.toString());
        if (BooleanUtil.isFalse(isMember)) {
            // 3.如果未点赞，可以点赞
            // 3.1 数据库点赞数 + 1
            boolean isSuccess = update().setSql("liked = liked + 1").eq("id", id).update();
            // 3.2 保存用户到redis的set集合
            if (isSuccess) {
                stringRedisTemplate.opsForSet().add(key, userId.toString());
            }
        } else {
            // 4.如果已点赞，取消点赞
            // 4.1 数据库点赞数 - 1
            boolean isSuccess = update().setSql("liked = liked - 1").eq("id", id).update();
            // 4.2 把用户从redis的set集合移除
            if (isSuccess) {
                stringRedisTemplate.opsForSet().remove(key, userId);
            }
        }
        return Result.ok();
    }

    private void queryBlogUser(Blog blog) {
        Long userId = blog.getUserId();
        User user = userService.getById(userId);
        blog.setName(user.getNickName());
        blog.setIcon(blog.getIcon());
    }
}
```



## 点赞排行榜

在探店笔记的详情页面，应该把给该笔记点赞的人显示出来，比如最早点赞的TOP5，形成点赞排行榜：

![image-20241130140306760](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130140306760.png)



### 案例：实现查询点赞排行榜的接口

需求：按照点赞时间先后排序，返回Top5的用户

![image-20241130140717387](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130140717387.png)



BlogController.java

```java
@GetMapping("/likes/{id}")
public Result queryBlogLikes(@PathVariable("id") Long id) {
    return blogService.queryBlogLikes(id);
}
```

IBlogService.java

```java
Result queryBlogLikes(Long id);
```

BlogServiceImpl.java

```java
private void isBlockLiked(Blog blog) {
    // 1.获取登录用户
    UserDTO user = UserHolder.getUser();
    if (user == null) {
        // 用户未登录，无需查询是否点赞
        return;
    }
    Long userId = user.getId();
    // 2.判断当前登录用户是否已经点赞
    String key = RedisConstants.BLOG_LIKED_KEY + blog.getId();
    Double score = stringRedisTemplate.opsForZSet().score(key, userId.toString());
    blog.setIsLike(score != null);
}

@Override
public Result likeBlog(Long id) {
    // 1.获取登录用户
    Long userId = UserHolder.getUser().getId();
    // 2.判断当前登录用户是否已经点赞
    String key = RedisConstants.BLOG_LIKED_KEY + id;
    Double score = stringRedisTemplate.opsForZSet().score(key, userId.toString());
    if (score == null) {
        // 3.如果未点赞，可以点赞
        // 3.1 数据库点赞数 + 1
        boolean isSuccess = update().setSql("liked = liked + 1").eq("id", id).update();
        // 3.2 保存用户到redis的sortedset集合 zadd key value score
        if (isSuccess) {
            stringRedisTemplate.opsForZSet().add(key, userId.toString(), System.currentTimeMillis());
        }
    } else {
        // 4.如果已点赞，取消点赞
        // 4.1 数据库点赞数 - 1
        boolean isSuccess = update().setSql("liked = liked - 1").eq("id", id).update();
        // 4.2 把用户从redis的set集合移除
        if (isSuccess) {
            stringRedisTemplate.opsForZSet().remove(key, userId);
        }
    }
    return Result.ok();
}

@Override
public Result queryBlogLikes(Long id) {
    String key = RedisConstants.BLOG_LIKED_KEY + id;
    // 1.查询Top5的点赞用户 zrange key 0 4
    Set<String> top5 = stringRedisTemplate.opsForZSet().range(key, 0, 4);
    if (top5 == null || top5.isEmpty()) {
        return Result.ok(Collections.emptyList());
    }
    // 2.解析出其中的用户id
    List<Long> ids = top5.stream().map(Long::valueOf).collect(Collectors.toList());
    String idStr = StrUtil.join(",", ids);
    // 3.根据用户id查询用户 where id in (5, 1) order by field(id, 5, 1)
    List<UserDTO> userDTOS = userService.query()
            .in("id", ids).last("order by field(id," + idStr + ")").list()
            .stream()
            .map(user -> BeanUtil.copyProperties(user, UserDTO.class))
            .collect(Collectors.toList());
    // 4.返回
    return Result.ok(userDTOS);
}
```



# 好友关注

## 关注和取关

在探店图文的详情页面中，可以关注发布笔记的作者：

![image-20241130144227294](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130144227294.png)



### 案例：实现关注和取关功能

**需求**：基于该表数据结构，实现两个接口：

①关注和取关接口

②判断是否关注的接口

关注是User之间的关系，是博主与粉丝的关系，数据库中有一张tb_follow表来标示：

![image-20241130144719057](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130144719057.png)

注意: 这里需要把主键修改为自增长，简化开发。



FollowController.java

```java
public class FollowController {

    @Resource
    private IFollowService followService;

    @PutMapping("/{id}/{isFollow}")
    public Result follow(@PathVariable("id") Long followUserId, @PathVariable("isFollow") Boolean isFollow) {
        return followService.follow(followUserId, isFollow);
    }

    @GetMapping("/or/not/{id}")
    public Result isFollow(@PathVariable("id") Long followUserId) {
        return followService.isFollow(followUserId);
    }
}
```

IFollowService.java

```java
public interface IFollowService extends IService<Follow> {

    Result follow(Long followUserId, Boolean isFollow);

    Result isFollow(Long followUserId);

}
```

FollowServiceImpl.java

```java
@Service
public class FollowServiceImpl extends ServiceImpl<FollowMapper, Follow> implements IFollowService {

    @Override
    public Result follow(Long followUserId, Boolean isFollow) {
        // 1.获取登录的用户
        Long userId = UserHolder.getUser().getId();
        // 2.判断到底是关注还是取关
        if (isFollow) {
            // 3.关注，新增数据
            Follow follow = new Follow();
            follow.setUserId(userId);
            follow.setFollowUserId(followUserId);
            save(follow);
        } else {
            // 4.取关，删除数据 delete from tb_follow where user_id = ? and follow_user_id = ?
            remove(new QueryWrapper<Follow>()
                    .eq("user_id", userId).eq("follow_user_id", followUserId));
        }
        return Result.ok();
    }

    @Override
    public Result isFollow(Long followUserId) {
        // 1.获取登录的用户
        Long userId = UserHolder.getUser().getId();
        // 2.查询是否关注 select count(*) from tb_follow where user_id = ? and follow_user_id = ?
        Integer count = query().eq("user_id", userId).eq("follow_user_id", followUserId).count();
        // 3.判断
        return Result.ok(count > 0);
    }
}
```



## 共同关注

点击博主头像，可以进入博主首页：

![image-20241130151215435](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130151215435.png)



### 博主个人主页

博主个人首页依赖两个接口：

1. 根据id查询user信息：

![image-20241130151250096](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130151250096.png)

2. 根据id查询博主的探店笔记：

![image-20241130151324494](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130151324494.png)



### 案例：实现共同关注功能

需求：利用Redis中恰当的数据结构，实现共同关注功能。在博主个人页面展示出当前用户与博主的共同好友。

![image-20241130151722259](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130151722259.png)



FollowController.java

```java
@RestController
@RequestMapping("/follow")
public class FollowController {

    @GetMapping("/common/{id}")
    public Result followCommons(@PathVariable("id") Long id) {
        return followService.followCommons(id);
    }
}
```

IFollowService.java

```java
public interface IFollowService extends IService<Follow> {
   Result followCommons(Long id);
}
```

FollowServiceImpl.java

```java
@Service
public class FollowServiceImpl extends ServiceImpl<FollowMapper, Follow> implements IFollowService {

    @Resource
    private StringRedisTemplate stringRedisTemplate;

    @Resource
    private IUserService userService;

    @Override
    public Result follow(Long followUserId, Boolean isFollow) {
        // 1.获取登录的用户
        Long userId = UserHolder.getUser().getId();
        // 2.判断到底是关注还是取关
        String key = "follows:" + userId;
        if (isFollow) {
            // 3.关注，新增数据
            Follow follow = new Follow();
            follow.setUserId(userId);
            follow.setFollowUserId(followUserId);
            boolean isSuccess = save(follow);
            if (isSuccess) {
                // 把关注用户的id，放入redis的set集合
                stringRedisTemplate.opsForSet().add(key, followUserId.toString());
            }
        } else {
            // 4.取关，删除数据 delete from tb_follow where user_id = ? and follow_user_id = ?
            boolean isSuccess = remove(new QueryWrapper<Follow>()
                    .eq("user_id", userId).eq("follow_user_id", followUserId));
            if (isSuccess) {
                // 把关注的用户id从redis集合中移除
                stringRedisTemplate.opsForSet().remove(key, followUserId);
            }
        }
        return Result.ok();
    }

    @Override
    public Result followCommons(Long id) {
        // 1.获取当前登录用户
        Long userId = UserHolder.getUser().getId();
        String key = "follows:" + userId;
        // 2.求交集
        String key2 = "follows:" + id;
        Set<String> intersect = stringRedisTemplate.opsForSet().intersect(key, key2);
        if (intersect == null || intersect.isEmpty()) {
            return Result.ok(Collections.emptyList());
        }
        // 3.解析id集合
        List<Long> ids = intersect.stream().map(Long::valueOf).collect(Collectors.toList());
        // 4.查询用户
        List<UserDTO> userDTOS = userService.listByIds(ids)
                .stream()
                .map(user -> BeanUtil.copyProperties(user, UserDTO.class))
                .collect(Collectors.toList());
        return Result.ok(userDTOS);
    }
}
```



## 关注推送

关注推送也叫做Feed流，直译为**投喂**。为用户持续的提供“沉浸式”的体验，通过无限下拉刷新获取新的信息。

![image-20241130154356235](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130154356235.png)



### Feed流模式

Feed流产品有两种常见模式：

- **Timeline**：不做内容筛选，简单的按照内容发布时间排序，常用于好友或关注。例如朋友圈
  - 优点：信息全面，不会有缺失。并且实现也相对简单
  - 缺点：信息噪音较多，用户不一定感兴趣，内容获取效率低

- **智能排序**：利用智能算法屏蔽掉违规的、用户不感兴趣的内容。推送用户感兴趣信息来吸引用户
  - 优点：投喂用户感兴趣信息，用户粘度很高，容易沉迷
  - 缺点：如果算法不精准，可能起到反作用

本例中的个人页面，是基于关注的好友来做Feed流，因此采用Timeline的模式。该模式的实现方案有三种：

1. 拉模式

2. 推模式

3. 推拉结合



#### Feed流的实现方案1

**拉模式**：也叫做读扩散

![image-20241130155245241](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130155245241.png)



#### Feed流的实现方案2

**推模式**：也叫做写扩散。

![image-20241130155402026](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130155402026.png)



#### Feed流的实现方案3

**推拉结合模式**：也叫做读写混合，兼具推和拉两种模式的优点。

![image-20241130155822808](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130155822808.png)

### Feed流的实现方案

![image-20241130155943897](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130155943897.png)



### 案例：基于推模式实现关注推送功能

**需求**：

①修改新增探店笔记的业务，在保存blog到数据库的同时，推送到粉丝的收件箱

②收件箱满足可以根据时间戳排序，必须用Redis的数据结构实现

③查询收件箱数据时，可以实现分页查询

![image-20241130160014823](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130160014823.png)

BlogController.java

```java
@PostMapping
public Result saveBlog(@RequestBody Blog blog) {
    return blogService.saveBlog(blog);
}
```

BlogServiceImpl.java

```java
@Override
public Result saveBlog(Blog blog) {
    // 1.获取登录用户
    UserDTO userDTO = UserHolder.getUser();
    blog.setUserId(userDTO.getId());
    // 2.保存探店笔记
    boolean isSuccess = save(blog);
    if (!isSuccess) {
        return Result.fail("新增笔记失败！");
    }
    // 3.查询笔记作者的所有粉丝 select * from tb_follow where follow_user_id = ?
    List<Follow> follows = followService.query().eq("follow_user_id", userDTO.getId()).list();
    // 4.推送笔记id给所有粉丝
    for (Follow follow : follows) {
        // 4.1 获取粉丝id
        Long userId = follow.getUserId();
        // 4.2 推送
        String key = "feed:" + userId;
        stringRedisTemplate.opsForZSet().add(key, blog.getId().toString(), System.currentTimeMillis());
    }
    // 5.返回id
    return Result.ok(blog.getId());
}
```



### Feed流的分页问题

Feed流中的数据会不断更新，所以数据的角标也在变化，因此不能采用传统的分页模式。

![image-20241130161650383](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130161650383.png)

### Feed流的滚动分页

![image-20241130161819877](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130161819877.png)



### 案例：实现关注推送页面的分页查询

需求：在个人主页的“关注”卡片中，查询并展示推送的Blog信息：

![image-20241130161924815](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241130161924815.png)



BlogController.java

```java
@GetMapping("/of/follow")
public Result queryBlogOfFollow(
        @RequestParam("lastId") Long max,
        @RequestParam(value = "offset", defaultValue = "0") Integer offset) {
    return blogService.queryBlogOfFollow(max, offset);
}
```

BlogServiceImpl.java

```java
@Override
public Result queryBlogOfFollow(Long max, Integer offset) {
    // 1.获取当前用户
    Long userId = UserHolder.getUser().getId();
    // 2.查询收件箱 zrevrangebyscore key max min limit offset count
    String key = RedisConstants.FEED_KEY + userId;
    Set<ZSetOperations.TypedTuple<String>> typedTuples = stringRedisTemplate.opsForZSet().reverseRangeByScoreWithScores(key, 0, max, offset, 2);
    // 3.非空判断
    if (typedTuples == null || typedTuples.isEmpty()) {
        return Result.ok();
    }
    // 4.解析数据：blogId、minTime（时间戳）、offset
    List<Long> ids = new ArrayList<>(typedTuples.size());
    long minTime = 0;
    int os = 1;
    for (ZSetOperations.TypedTuple<String> tuple : typedTuples) {
        // 4.1 获取id
        ids.add(Long.valueOf(tuple.getValue()));
        // 4.2 获取分数（时间戳）
        long time = tuple.getScore().longValue();
        if (time == minTime) {
            os++;
        } else {
            minTime = time;
            os = 1;
        }
    }
    // 5.根据id查询blog
    String idStr = StrUtil.join(",", ids);
    List<Blog> blogs = query().in("id", ids).last("order by field(id," + idStr + ")").list();

    for (Blog blog : blogs) {
        // 5.1 查询blog有关的用户
        queryBlogUser(blog);
        // 5.2 查询blog是否被点赞
        isBlockLiked(blog);
    }

    // 6.封装并返回
    ScrollResult r = new ScrollResult();
    r.setList(blogs);
    r.setOffset(os);
    r.setMinTime(minTime);

    return Result.ok(r);
}
```



# 附近商户

## GEO数据结构

GEO就是Geolocation的简写形式，代表地理坐标。Redis在3.2版本中加入了对GEO的支持，允许存储地理坐标信息，帮助我们根据经纬度来检索数据。常见的命令有：

[GEOADD](https://redis.io/commands/geoadd)：添加一个地理空间信息，包含：经度（longitude）、纬度（latitude）、值（member）

[GEODIST](https://redis.io/commands/geodist)：计算指定的两个点之间的距离并返回

[GEOHASH](https://redis.io/commands/geohash)：将指定member的坐标转为hash字符串形式并返回

[GEOPOS](https://redis.io/commands/geopos)：返回指定member的坐标

[GEORADIUS](https://redis.io/commands/georadius)：指定圆心、半径，找到该圆内包含的所有member，并按照与圆心之间的距离排序后返回。6.2以后已废弃

[GEOSEARCH](https://redis.io/commands/geosearch)：在指定范围内搜索member，并按照与指定点之间的距离排序后返回。范围可以是圆形或矩形。6.2.新功能

[GEOSEARCHSTORE](https://redis.io/commands/geosearchstore)：与GEOSEARCH功能一致，不过可以把结果存储到一个指定的key。 6.2.新功能



## 案例：练习Redis的GEO功能

需求：

1. 添加下面几条数据：
   - 北京南站（ 116.378248 39.865275 ）
   - 北京站（ 116.42803 39.903738 ）
   - 北京西站（ 116.322287 39.893729 ）

2. 计算北京西站到北京站的距离

3. 搜索天安门（ 116.397904 39.909005 ）附近10km内的所有火车站，并按照距离升序排序



```java
@Test
void loadShopData() {
    // 1.查询店铺信息
    List<Shop> list = shopService.list();
    // 2.把店铺分组，按照typeId分组，typeId一致的放到一个集合
    Map<Long, List<Shop>> map = list.stream().collect(Collectors.groupingBy(Shop::getTypeId));
    // 3.分批完成写入Redis
    for (Map.Entry<Long, List<Shop>> entry : map.entrySet()) {
        // 3.1 获取类型id
        Long typeId = entry.getKey();
        String key = RedisConstants.SHOP_GEO_KEY + typeId;
        // 3.2 获取同类型的店铺集合
        List<Shop> value = entry.getValue();
        List<RedisGeoCommands.GeoLocation<String>> locations = new ArrayList<>(value.size());
        // 3.3 写入Redis geoadd key 经度 纬度 member
        for (Shop shop : value) {
            // stringRedisTemplate.opsForGeo().add(key, new Point(shop.getX(), shop.getY()), shop.getId().toString());
            locations.add(new RedisGeoCommands.GeoLocation<>(shop.getId().toString(), new Point(shop.getX(), shop.getY())));
        }
        stringRedisTemplate.opsForGeo().add(key, locations);
    }
}
```



## 附近商户搜索

在首页中点击某个频道，即可看到频道下的商户：

![image-20241201162725585](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241201162725585.png)



按照商户类型做分组，类型相同的商户作为同一组，以typeId为key存入同一个GEO集合中即可

![image-20241201163256835](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241201163256835.png)

SpringDataRedis的2.3.9版本并不支持Redis 6.2提供的GEOSEARCH命令，因此我们需要提示其版本，修改自己的POM文件，内容如下：

```xml
<dependency>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-data-redis</artifactId>
    <exclusions>
        <exclusion>
            <groupId>org.springframework.data</groupId>
            <artifactId>spring-data-redis</artifactId>
        </exclusion>
        <exclusion>
            <artifactId>lettuce-core</artifactId>
            <groupId>io.lettuce</groupId>
        </exclusion>
    </exclusions>
</dependency>
<dependency>
    <groupId>org.springframework.data</groupId>
    <artifactId>spring-data-redis</artifactId>
    <version>2.6.2</version>
</dependency>
<dependency>
    <artifactId>lettuce-core</artifactId>
    <groupId>io.lettuce</groupId>
    <version>6.1.6.RELEASE</version>
</dependency>
```

ShopController.java

```java
@GetMapping("/of/type")
public Result queryShopByType(
        @RequestParam("typeId") Integer typeId,
        @RequestParam(value = "current", defaultValue = "1") Integer current,
        @RequestParam(value = "x", required = false) Double x,
        @RequestParam(value = "y", required = false) Double y
) {
    return shopService.queryShopByType(typeId, current, x, y);
}
```

ShopServiceImpl.java

```java
@Override
public Result queryShopByType(Integer typeId, Integer current, Double x, Double y) {
    // 1.判断是否需要根据坐标查询
    if (x == null || y == null) {
        // 不需要坐标查询，按照数据库查询
        Page<Shop> page = query().eq("type_id", typeId).page(new Page<>(current, SystemConstants.DEFAULT_PAGE_SIZE));
        // 返回数据
        return Result.ok(page.getRecords());
    }

    // 2.计算分页参数
    int from = (current - 1) * SystemConstants.DEFAULT_PAGE_SIZE;
    int end = current * SystemConstants.DEFAULT_PAGE_SIZE;

    // 3.查询redis，按照距离排序、分页。结果：shopId、distance
    // geosearch bylonlat xy byradius 10 withdistance
    String key = RedisConstants.SHOP_GEO_KEY + typeId;
    GeoResults<RedisGeoCommands.GeoLocation<String>> results = stringRedisTemplate.opsForGeo().
            search(
                    key,
                    GeoReference.fromCoordinate(x, y),
                    new Distance(5000),
                    RedisGeoCommands.GeoSearchCommandArgs.newGeoSearchArgs().includeDistance().limit(end)
            );

    // 4.解析出id
    if (results == null) {
        return Result.ok(Collections.emptyList());
    }
    List<GeoResult<RedisGeoCommands.GeoLocation<String>>> list = results.getContent();
    if (list.size() <= from) {
        // 没有下一页，结束
        return Result.ok(Collections.emptyList());
    }
    // 4.1 截取from到end的部分
    List<Long> ids = new ArrayList<>(list.size());
    Map<String, Distance> distanceMap = new HashMap<>(list.size());
    list.stream()
            .skip(from)
            .forEach(result -> {
                // 4.2 获取店铺id
                String shopIdStr = result.getContent().getName();
                ids.add(Long.valueOf(shopIdStr));
                // 4.3 获取距离
                Distance distance = result.getDistance();
                distanceMap.put(shopIdStr, distance);
            });
    // 5.根据id查询shop
    String idStr = StrUtil.join(",", ids);
    List<Shop> shops = query().in("id", ids).last("order by field(id" + idStr + ")").list();
    for (Shop shop : shops) {
        shop.setDistance(distanceMap.get(shop.getId().toString()).getValue());
    }
    // 6.返回
    return Result.ok(shops);
}
```



# 用户签到

## BitMap用法

假如我们用一张表来存储用户签到信息，其结构应该如下：

![image-20241201175423170](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241201175423170.png)

假如有1000万用户，平均每人每年签到次数为10次，则这张表一年的数据量为 1亿条

每签到一次需要使用（8 + 8 + 1 + 1 + 3 + 1）共22 字节的内存，一个月则最多需要600多字节

![image-20241201175443064](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241201175443064.png)

我们按月来统计用户签到信息，签到记录为1，未签到则记录为0。

![image-20241201180042016](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241201180042016.png)

把每一个bit位对应当月的每一天，形成了映射关系。用0和1标示业务状态，这种思路就称为**位图（BitMap）。**

**Redis**中是利用string类型数据结构实现**BitMap**，因此最大上限是512M，转换为bit则是 2^32个bit位。

BitMap的操作命令有：

[SETBIT](https://redis.io/commands/setbit)：向指定位置（offset）存入一个0或1

[GETBIT](https://redis.io/commands/getbit) ：获取指定位置（offset）的bit值

[BITCOUNT](https://redis.io/commands/bitcount) ：统计BitMap中值为1的bit位的数量

[BITFIELD](https://redis.io/commands/bitfield) ：操作（查询、修改、自增）BitMap中bit数组中的指定位置（offset）的值

[BITFIELD_RO](https://redis.io/commands/bitfield_ro) ：获取BitMap中bit数组，并以十进制形式返回

[BITOP](https://redis.io/commands/bitop) ：将多个BitMap的结果做位运算（与 、或、异或）

[BITPOS](https://redis.io/commands/bitpos) ：查找bit数组中指定范围内第一个0或1出现的位置





## 签到功能

### 案例：签到功能

需求：实现签到接口，将当前用户当天签到信息保存到Redis中

|          | **说明**   |
| -------- | ---------- |
| 请求方式 | Post       |
| 请求路径 | /user/sign |
| 请求参数 | 无         |
| 返回值   | 无         |

提示：因为BitMap底层是基于String数据结构，因此其操作也都封装在字符串相关操作中了。

![image-20241201195039399](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241201195039399.png)



UserServiceImpl.java

```java
@Override
public Result sign() {
    // 1.获取当前登录用户
    Long userId = UserHolder.getUser().getId();

    // 2.获取日期
    LocalDateTime now = LocalDateTime.now();

    // 3.拼接key
    String keySuffix = now.format(DateTimeFormatter.ofPattern(":yyyyMM"));
    String key = RedisConstants.USER_SIGN_KEY + userId + keySuffix;

    // 4.获取今天是本月的第几天
    int dayOfMonth = now.getDayOfMonth();

    // 5.写入Redis SETBIT key offset 1
    stringRedisTemplate.opsForValue().setBit(key, dayOfMonth - 1, true);

    return Result.ok();
}
```



## 签到统计

**问题1**：什么叫做连续签到天数？

从最后一次签到开始向前统计，直到遇到第一次未签到为止，计算总的签到次数，就是连续签到天数。

![image-20241201200512204](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241201200512204.png)

**问题2**：如何得到本月到今天为止的所有签到数据？

 BITFIELD key GET u[dayOfMonth] 0

**问题3**：如何从后向前遍历每个bit位？

与 1 做与运算，就能得到最后一个bit位。

随后右移1位，下一个bit位就成为了最后一个bit位。



### 案例：实现签到统计功能

需求：实现下面接口，统计当前用户截止当前时间在本月的连续签到天数

|          | **说明**         |
| -------- | ---------------- |
| 请求方式 | GET              |
| 请求路径 | /user/sign/count |
| 请求参数 | 无               |
| 返回值   | 连续签到天数     |



UserServiceImpl.java

```java
@Override
public Result signCount() {
    // 1.获取当前登录用户
    Long userId = UserHolder.getUser().getId();

    // 2.获取日期
    LocalDateTime now = LocalDateTime.now();

    // 3.拼接key
    String keySuffix = now.format(DateTimeFormatter.ofPattern(":yyyyMM"));
    String key = RedisConstants.USER_SIGN_KEY + userId + keySuffix;

    // 4.获取今天是本月的第几天
    int dayOfMonth = now.getDayOfMonth();

    // 5.获取本月截止今天为止的所有的签到记录，返回的是一个十进制的数字
    List<Long> results = stringRedisTemplate.opsForValue().bitField(
            key,
            BitFieldSubCommands.create().
                    get(BitFieldSubCommands.BitFieldType.unsigned(dayOfMonth)).valueAt(0));
    if (results == null || results.isEmpty()) {
        return Result.ok(0);
    }
    Long num = results.get(0);
    if (num == null || num == 0) {
        return Result.ok(0);
    }

    // 6.循环遍历
    int count = 0;
    while (true) {
        // 7.让这个数字与1做与运算，得到数字的最后一个bit位
        // 8.判断这个bit位是否为0
        if ((num & 1) == 0) {
            // 8.1 如果为0，说明未签到，结束
            break;
        } else {
            // 8.2 如果不为0，说明已签到，计数器加1
            count++;
        }

        // 9.把数字右移移位，抛弃最后一个bit位，继续下一个bit位
        num >>>= 1;
    }

    return Result.ok(count);
}
```



# UV统计

## HyperLogLog用法

首先我们搞懂两个概念：

- **UV**：全称**U**nique **V**isitor，也叫独立访客量，是指通过互联网访问、浏览这个网页的自然人。1天内同一个用户多次访问该网站，只记录1次。

- **PV**：全称**P**age **V**iew，也叫页面访问量或点击量，用户每访问网站的一个页面，记录1次PV，用户多次打开页面，则记录多次PV。往往用来衡量网站的流量。

UV统计在服务端做会比较麻烦，因为要判断该用户是否已经统计过了，需要将统计过的用户信息保存。但是如果每个访问的用户都保存到Redis中，数据量会非常恐怖。



Hyperloglog(HLL)是从Loglog算法派生的概率算法，用于确定非常大的集合的基数，而不需要存储其所有值。相关算法原理大家可以参考：[https://juejin.cn/post/6844903785744056333#heading-0](https://juejin.cn/post/6844903785744056333)

Redis中的HLL是基于string结构实现的，单个HLL的内存永远小于16kb，内存占用低的令人发指！作为代价，其测量结果是概率性的，有小于0.81％的误差。不过对于UV统计来说，这完全可以忽略。

![image-20241201203440301](https://picgo-store-imafes.oss-cn-wuhan-lr.aliyuncs.com/ob/image-20241201203440301.png)



## 实现UV统计

我们直接利用单元测试，向HyperLogLog中添加100万条数据，看看内存占用和统计效果如何：

```java
@Test
void testHyperLogLog() {
    // 准备数组，装用户数据
    String[] users = new String[1000];
    // 数组角标
    int index = 0;
    for (int i = 1; i <= 1000000; i++) {
        // 赋值
        users[index++] = "user_" + i;
        // 每1000条发送一次
        if (i % 1000 == 0) {
            index = 0;
            stringRedisTemplate.opsForHyperLogLog().add("hll1", users);
        }
    }
    // 统计数量
    Long size = stringRedisTemplate.opsForHyperLogLog().size("hll1");
    System.out.println("size = " + size);
}
```



HyperLogLog的作用：

- 做海量数据的统计工作

HyperLogLog的优点：

- 内存占用极低

- 性能非常好

HyperLogLog的缺点：

- 有一定的误差