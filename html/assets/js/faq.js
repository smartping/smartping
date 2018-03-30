/**
 * @name      FAQ Plugin
 * @author    Rod Howard
 * @url       http://goideate.com
 * @date      April 28, 2012
 * @license   GNU/GPL Version 3
 *
 *
 * Example:
 *
 *   $(function() {
 *     $("element, #id, .class").technify({
 *       // #{EDIT-HERE}# Your run-time options go here ...
 *     });
 *   });
 */
/**
 * Create an anonymous function to avoid library conflicts
 */
(function($) {
    /**
     * Add our plugin to the jQuery.fn object
     */
    $.fn.goFaq = function(options) {
        /**
         * Define some default settings
         */
        var defaults = {
        	enableSearch: true,
        	enableToc: true,
        	enableStyling: true,
            //numberHtml: '{{#}}.',
            numberHtml: '<div class="faq-number">{{#}}</div>'
        };
        /**
         * Merge the runtime options with the default settings
         */
        var options = $.extend({}, defaults, options);
        /**
         * Iterate through the collection of elements and
         * return the object to preserve method chaining
         */
        return this.each(function(i) {
            /**
             * Wrap the current element in an instance of jQuery
             */
            var $this = $(this);
            
            var $container = $this.wrap ('<div class="faq-container"></div>');
            
            $this.addClass ('faq-list');
            
            if (options.enableSearch) {            	
	            var $form = generateSearchForm ();
	            $form.insertBefore ($this);
            }
            
            if (options.enableToc) {
	            var $toc = generateToc ($this);
	            $toc.insertBefore ($this);
	        }
            			
            
            var $empty = generateEmptySearch ();
            $empty.appendTo ($container);
            
            generateAnswerNumbers ($this);
            
        });
        
        function search (e) {
			var el, container, filter, count, pattern, container, answers, toc, tocs, empty;
			
			el = $(this);
			container = el.parents ('.faq-container');
			filter = el.val ();
			toc = container.find ('.faq-toc');
			answers = container.find ('.faq-list').find ('li');
			tocs = container.find ('.faq-toc').find('li');
			empty = container.find ('.faq-empty');
			pattern = new RegExp (filter, 'i');
			
			answers.hide ();
			tocs.hide ();
			
			$.grep (answers.find ('.faq-text'), function (input) {
				if (pattern.test ($(input).text ())) {
					$(input).parents ('li').show ();
					
					var index = $(input).parents ('li').index ();				
					tocs.eq (index).show ();				
				}			
			});	
			
			if (!answers.is (':visible')) {
				empty.show ();
				toc.hide ();
			} else {
				empty.hide ();
				toc.show ();
			}
		}
        
		
		function generateEmptySearch () {
			var $empty = $('<div>', { 'class': 'faq-empty' });
			
			return $empty.html ('Nothing Found');
		}
        
        function generateSearchForm () {
        	
        	var $form = $('<form>', { 'class': 'faq-search' });
        	var $input = $('<input>', { 'type': 'text', 'name': 'search', 'placeholder': 'Search by Keyword' });
        	
        	$input.appendTo ($form);
        	
        	$input.bind ('keyup', search)
        	
        	return $form;
        }
        
        function generateAnswerNumbers ($list) {
        	$list.find ('li').each (function (i) {
        		var id = parseInt (i+1);
        		
        		
        		$(this).wrapInner ('<div class="faq-text"></div>');
        		
        		
            	if (options.enableStyling) {
					var icon = '<div class="faq-icon">' + options.numberHtml + '</div>';
	        		
					icon = icon.replace ('{{#}}', id);
					$(this).prepend (icon);
				}
        	});
        }
        
        function generateToc ($list) {
        	var html = '<ol>';	
        	
			$list.find ('li').each (function (i) {				
				var id = parseInt (i+1);							
				html += '<li>' + id + '. <a href="#faq-' + id + '">' + $(this).find ('h4').text () + '</a></li>';								
				$(this).attr ('id', 'faq-' + id);				
			});	
					
			html += '</ol>';
			
        	var $toc = $('<div>', { 'class': 'faq-toc' });
        	
        	$toc.html (html);
        	
        	return $toc;
        	
        }
    };
})(jQuery);